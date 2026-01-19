package model

import (
	"time"
)

// TableName{{.Web.R.SingularName}} defines the physical table name for the {{.Web.R.GORMModel}} model.
const TableName{{.Web.R.SingularName}} = "{{.Web.BinaryName | extractProjectPrefix}}_{{.Web.R.SingularLower}}"

// {{.Web.R.GORMModel}} represents the data model for the {{.Web.R.SingularLower}} resource.
// It maps to the "{{.Web.BinaryName | extractProjectPrefix}}_{{.Web.R.SingularLower}}" table in the database.
type {{.Web.R.GORMModel}} struct {
    // ID is the primary key of the record.
    ID int64 `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`

    // {{.Web.R.Last.SingularName}}ID is the unique identifier for the resource.
    {{.Web.R.Last.SingularName}}ID string `gorm:"column:{{.Web.R.Last.SingularLower}}_id;not null;comment:Unique resource ID" json:"{{.Web.R.Last.SingularLower}}_id"`

    // CreatedAt is the timestamp when the resource was created.
    CreatedAt time.Time `gorm:"column:created_at;not null;default:current_timestamp;comment:Creation timestamp" json:"createdAt"`

    // UpdatedAt is the timestamp when the resource was last modified.
    UpdatedAt time.Time `gorm:"column:updated_at;not null;default:current_timestamp;comment:Last modification timestamp" json:"updatedAt"`
}

// TableName returns the physical table name for the {{.Web.R.GORMModel}} model.
func (*{{.Web.R.GORMModel}}) TableName() string {
	return TableName{{.Web.R.SingularName}}
}
