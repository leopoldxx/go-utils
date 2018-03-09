package instance

import (
	"sync"

	"github.com/leopoldxx/go-utils/mysql"
	"github.com/spf13/viper"
)

/*
[mysql]
	connection = "test:test@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=true&loc=Asia%2FShanghai"
	maxConnsCount = 100
	maxIdleConnsCount = 50
*/
// config keys
const (
	mysqlConnection    = "mysql.connection"
	mysqlMaxConnsCount = "mysql.maxConnsCount"
	mysqlMaxIdleConns  = "mysql.maxIdleConnsCount"
)

var (
	mysqlClient *mysql.Client
	mysqlOnce   sync.Once
)

func init() {
	initDaoDefaultConfigs()
}

func initDaoDefaultConfigs() {
	viper.SetDefault(mysqlConnection, "root@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=true&loc=Asia%2FShanghai")
	viper.SetDefault(mysqlMaxConnsCount, 100)
	viper.SetDefault(mysqlMaxIdleConns, 50)
}

// GetMySQLClient create a mysql backend storage Client
func GetMySQLClient() *mysql.Client {
	mysqlOnce.Do(func() {
		mysqlClient, _ = mysql.New(viper.GetString(mysqlConnection),
			mysql.WithMaxConnsCount(viper.GetInt(mysqlMaxConnsCount)),
			mysql.WithMaxIdleConnsCount(viper.GetInt(mysqlMaxIdleConns)))
		if mysqlClient == nil {
			panic("connect mysql failed")
		}
	})
	return mysqlClient
}
