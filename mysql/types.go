package mysql

// WhereClause for sql
type WhereClause map[Field]Value

// Field name of a row field
type Field string

// Value of a msyql column
type Value interface{}

// Fields for mysql tables
const (
	FieldID         = Field("id")
	FieldUUID       = Field("uuid")
	FieldCreateTime = Field("create_time")
	FieldUpdateTime = Field("last_update_time")
	FieldName       = Field("name")
	FieldIsDeleted  = Field("is_deleted")
	FieldDeleteTime = Field("deleted_time")
	FieldUser       = Field("user")
	FieldType       = Field("typ")
)
