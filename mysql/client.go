package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/VividCortex/mysqlerr"
	gomysql "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/leopoldxx/go-utils/errors"
	"github.com/leopoldxx/go-utils/trace"
	uuid "github.com/satori/go.uuid"
)

const (
	maxPlaceHolder = 65535
)

type options struct {
	// for client initialization
	maxOpenConnsCount int
	maxIdleConnsCount int
	// for operation
	extra string
}

// Option for MySQL Client
type Option func(opt *options)

// WithMaxConnsCount set max conns count for MySQL Client
func WithMaxConnsCount(count int) Option {
	return func(opt *options) {
		opt.maxOpenConnsCount = count
	}
}

// WithMaxIdleConnsCount set max idle conns count for MySQL Client
func WithMaxIdleConnsCount(count int) Option {
	return func(opt *options) {
		opt.maxIdleConnsCount = count
	}
}

// New will create a mysql-backend client
func New(addr string, ops ...Option) (*Client, error) {
	opts := &options{
		maxOpenConnsCount: 10,
		maxIdleConnsCount: 10,
	}
	for _, op := range ops {
		op(opts)
	}

	db, err := sqlx.Open("mysql", addr)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(opts.maxOpenConnsCount)
	db.SetMaxIdleConns(opts.maxIdleConnsCount)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &Client{db: db}, nil
}

// Client for mysql db
type Client struct {
	db *sqlx.DB
}

// Close the db behind the mysql client
func (cli *Client) Close() error {
	if cli == nil || cli.db == nil {
		return nil
	}
	return cli.db.Close()
}

// WithExtra set extra sql statements
func WithExtra(extra string) Option {
	return func(opts *options) {
		opts.extra = extra
	}
}

// TransactionHandler is a wrapper function for mysql transcations
func TransactionHandler(ctx context.Context, db *sqlx.DB, txFunc func(*sqlx.Tx) error) (err error) {
	tracer := trace.GetTraceFromContext(ctx)
	var tx *sqlx.Tx
	tx, err = db.Beginx()
	tracer.Infof("tx starting...")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			tracer.Errorf("tx rollbacked: %s", err)
		} else {
			tx.Commit()
			tracer.Info("tx committed")
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			tracer.Errorf("tx panic: %v", r)
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("tx: %v", r)
			}
		}
	}()

	if err = txFunc(tx); err != nil {
		return processErrors(err)
	}
	return nil
}

func formatWhereClauseParameters(whereClauses []WhereClause) (string, []interface{}) {
	var (
		whereClauseText   []string
		whereClauseValues []interface{}
	)
	for _, whereClause := range whereClauses {
		if len(whereClause) == 0 {
			continue
		}
		var (
			whereFieldText   []string
			whereFieldsValue []interface{}
		)
		for k, v := range whereClause {
			value := reflect.ValueOf(v)
			if value.Kind() == reflect.Slice {
				slen := value.Len()
				if slen > 0 {
					placeholder := []string{}
					valueholder := []interface{}{}
					for i := 0; i < slen; i++ {
						placeholder = append(placeholder, "?")
						valueholder = append(valueholder,
							value.Index(i).Interface())
					}
					whereFieldText = append(whereFieldText, fmt.Sprintf("%s IN (%s)",
						string(k), strings.Join(placeholder, ",")))
					whereFieldsValue = append(whereFieldsValue, valueholder...)
				}
			} else {
				if fv, ok := v.(Field); ok {
					whereFieldText = append(whereFieldText, fmt.Sprintf("%s=%s", string(k), string(fv)))
				} else {
					whereFieldText = append(whereFieldText, fmt.Sprintf("%s=?", string(k)))
					whereFieldsValue = append(whereFieldsValue, v)
				}
			}
		}
		whereClauseText = append(whereClauseText, fmt.Sprintf("(%s)", strings.Join(whereFieldText, " AND ")))
		whereClauseValues = append(whereClauseValues, whereFieldsValue...)
	}
	return strings.Join(whereClauseText, " OR "), whereClauseValues
}

func formatSelectParameters(table string, rowFields []Field, whereClause []WhereClause) (string, []interface{}, error) {
	var sqlTpl = `SELECT %s FROM %s %s`
	fields := []string{}
	for i := 0; i < len(rowFields); i++ {
		fields = append(fields, string(rowFields[i]))
	}
	fieldText := fmt.Sprintf("%s", strings.Join(fields, ","))
	whereClauseText, whereFieldsValue := formatWhereClauseParameters(whereClause)

	return fmt.Sprintf(sqlTpl, fieldText, table, func() string {
		if len(whereClauseText) == 0 {
			return ""
		}
		return fmt.Sprintf(" WHERE %s", whereClauseText)
	}()), whereFieldsValue, nil
}

// SelectRows is a util function to select some rows from a table
func SelectRows(ctx context.Context, db *sqlx.DB, tx *sqlx.Tx, table string, fields []Field, whereClause []WhereClause, result interface{}, ops ...Option) error {
	opts := &options{}
	for _, op := range ops {
		op(opts)
	}
	tracer := trace.GetTraceFromContext(ctx)
	sqlTpl, fieldsValue, err := formatSelectParameters(table, fields, whereClause)
	if err != nil {
		tracer.Errorf("failed to format select sql: %s %s", table, err)
		return err
	}
	if len(opts.extra) > 0 {
		sqlTpl = sqlTpl + " " + opts.extra
	}

	if db != nil {
		err = db.Select(result, sqlTpl, fieldsValue...)
	} else if tx != nil {
		err = tx.Select(result, sqlTpl, fieldsValue...)
	} else {
		return errors.NewBadRequestError("invalid db handler")
	}

	if err != nil {
		if isNoRowsError(err) {
			tracer.Errorf("not found any rows by %v", fields)
			return errors.NewNotFoundError(fmt.Sprintf("%v", fields))
		}
		tracer.Errorf("failed to get data: %v", err)
		return err
	}
	return nil
}

func formatInsertParameters(table string, rowFields []Field, rowValues [][]Value) (string, []interface{}, error) {
	var (
		sqlTpl = `INSERT INTO %s (%s) VALUES %s`
	)
	var (
		fieldValueHolders []string
		fieldValues       []interface{}
	)
	fields := []string{}
	for i := 0; i < len(rowFields); i++ {
		fields = append(fields, string(rowFields[i]))
	}
	fieldText := fmt.Sprintf("%s", strings.Join(fields, ","))
	for i := 0; i < len(rowValues); i++ {
		holders := []string{}
		values := []interface{}{}
		// validate values and their types
		for j := 0; j < len(rowValues[i]); j++ {
			holders = append(holders, "?")
			values = append(values, rowValues[i][j])
		}
		fieldValueHolders = append(fieldValueHolders, fmt.Sprintf("(%s)", strings.Join(holders, ",")))
		fieldValues = append(fieldValues, values...)
	}
	sqlTpl = fmt.Sprintf(sqlTpl, table, fieldText, strings.Join(fieldValueHolders, ","))

	return sqlTpl, fieldValues, nil
}

// InsertRows is a util function to insert some rows into a table
func InsertRows(ctx context.Context, db *sqlx.DB, tx *sqlx.Tx, table string, rowFields []Field, rowValues [][]Value, ops ...Option) (int64, error) {
	var count int64

	if len(rowValues) > 0 {
		maxBatchCount := maxPlaceHolder / len(rowValues[0])
		for len(rowValues) > 0 {
			batchCount := maxBatchCount
			if len(rowValues) < batchCount {
				batchCount = len(rowValues)
			}
			c, err := insertRows(ctx, db, tx, table, rowFields, rowValues[:batchCount])
			if err != nil {
				return 0, err
			}
			rowValues = rowValues[batchCount:]
			count += c
		}
	}
	return count, nil
}
func insertRows(ctx context.Context, db *sqlx.DB, tx *sqlx.Tx, table string, rowFields []Field, rowValues [][]Value, ops ...Option) (int64, error) {
	opts := &options{}
	for _, op := range ops {
		op(opts)
	}
	tracer := trace.GetTraceFromContext(ctx)
	sqlTpl, fieldValues, err := formatInsertParameters(table, rowFields, rowValues)
	if err != nil {
		tracer.Errorf("failed to format insert sql: %s %s", table, err)
		return 0, err
	}
	if len(opts.extra) > 0 {
		sqlTpl = sqlTpl + " " + opts.extra
	}

	var result sql.Result
	if db != nil {
		result, err = db.Exec(sqlTpl, fieldValues...)
	} else if tx != nil {
		result, err = tx.Exec(sqlTpl, fieldValues...)
	} else {
		return 0, errors.NewBadRequestError("invalid db handler")
	}
	if err != nil {
		tracer.Errorf("failed to insert table %s: %s", table, err)
		return 0, processErrors(err)
	}
	num, _ := result.RowsAffected()
	tracer.Infof("insert table %s #%d rows successfully", table, num)
	return num, nil
}

func formatUpdateParameters(table string, values map[Field]Value, whereClause []WhereClause) (string, []interface{}, error) {
	var sqlTpl = `UPDATE %s SET %s %s`
	var fieldValues []interface{}
	var fields []string
	for k, v := range values {
		fields = append(fields, fmt.Sprintf("%s=?", string(k)))
		fieldValues = append(fieldValues, v)
	}
	fieldText := strings.Join(fields, ",")
	whereClauseText, whereFieldsValue := formatWhereClauseParameters(whereClause)

	return fmt.Sprintf(sqlTpl, table, fieldText, func() string {
		if len(whereClauseText) == 0 {
			return ""
		}
		return fmt.Sprintf(" WHERE %s", whereClauseText)
	}()), append(fieldValues, whereFieldsValue...), nil
}

// UpdateRows is a util function to update rows values in a table
func UpdateRows(ctx context.Context, db *sqlx.DB, tx *sqlx.Tx, table string, values map[Field]Value, whereClause []WhereClause) (int64, error) {
	tracer := trace.GetTraceFromContext(ctx)
	sqlTpl, fieldValues, err := formatUpdateParameters(table, values, whereClause)
	if err != nil {
		tracer.Errorf("failed to format update sql: %s %s", table, err)
		return 0, err
	}

	var result sql.Result
	if db != nil {
		result, err = db.Exec(sqlTpl, fieldValues...)
	} else if tx != nil {
		result, err = tx.Exec(sqlTpl, fieldValues...)
	} else {
		return 0, errors.NewBadRequestError("invalid db handler")
	}
	if err != nil {
		tracer.Errorf("failed to update table %s: %s", table, err)
		return 0, processErrors(err)
	}
	num, _ := result.RowsAffected()
	tracer.Infof("update table %s #%d rows successfully", table, num)
	return num, nil
}

func formatDeleteParameters(table string, whereClause []WhereClause) (string, []interface{}, error) {
	var sqlTpl = `DELETE FROM %s %s`
	whereClauseText, whereFieldsValue := formatWhereClauseParameters(whereClause)

	return fmt.Sprintf(sqlTpl, table, func() string {
		if len(whereClauseText) == 0 {
			return ""
		}
		return fmt.Sprintf(" WHERE %s", whereClauseText)
	}()), whereFieldsValue, nil
}

// DeleteRows is a util function to delete rows from a table
func DeleteRows(ctx context.Context, db *sqlx.DB, tx *sqlx.Tx, table string, whereClause []WhereClause, ops ...Option) (int64, error) {
	opts := &options{}
	for _, op := range ops {
		op(opts)
	}
	tracer := trace.GetTraceFromContext(ctx)
	sqlTpl, fieldValues, err := formatDeleteParameters(table, whereClause)
	if err != nil {
		tracer.Errorf("failed to format delete sql: %s %s", table, err)
		return 0, err
	}

	if len(opts.extra) > 0 {
		sqlTpl = sqlTpl + " " + opts.extra
	}

	var result sql.Result
	if db != nil {
		result, err = db.Exec(sqlTpl, fieldValues...)
	} else if tx != nil {
		result, err = tx.Exec(sqlTpl, fieldValues...)
	} else {
		return 0, errors.NewBadRequestError("invalid db handler")
	}
	if err != nil {
		tracer.Errorf("failed to delete table %s: %s", table, err)
		return 0, processErrors(err)
	}
	num, _ := result.RowsAffected()
	tracer.Infof("delete table %s #%d rows successfully", table, num)
	return num, nil
}

func processErrors(err error) error {
	switch err {
	case sql.ErrNoRows:
		return errors.NewNotFoundError(err.Error())
	}
	if driverErr, ok := err.(*gomysql.MySQLError); ok {
		switch driverErr.Number {
		case mysqlerr.ER_DUP_ENTRY:
			return errors.NewConflictError(err.Error())
		}
		return driverErr
	}
	return err
}
func isNoRowsError(err error) bool {
	if err == sql.ErrNoRows {
		return true
	}
	return false
}

func genUUID() string {
	id := ""
	uid, err := uuid.NewV4()
	if err == nil {
		//strings.Trim(id)
		id = strings.Replace(uid.String(), "-", "", -1)
	} else {
		id = strconv.FormatInt(time.Now().Unix(), 36)
	}
	return id
}
