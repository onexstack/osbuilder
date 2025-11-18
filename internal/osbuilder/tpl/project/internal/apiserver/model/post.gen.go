package model

import (
	"time"
)

const TableName{{.Web.R.GORMModel}} = "{{.Web.BinaryName | extractProjectPrefix}}_{{.Web.R.SingularLower}}"

// {{.Web.R.GORMModel}} mapped from table <{{.Web.R.SingularLower}}>
type {{.Web.R.GORMModel}} struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	{{.Web.R.Last.SingularName}}ID    string    `gorm:"column:{{.Web.R.Last.SingularLowerFirst}}ID;not null;comment:资源唯一 ID" json:"{{.Web.R.Last.SingularLowerFirst}}ID"` // 资源唯一 ID
	CreatedAt time.Time `gorm:"column:createdAt;not null;default:current_timestamp;comment:资源创建时间" json:"createdAt"`   // 资源创建时间
	UpdatedAt time.Time `gorm:"column:updatedAt;not null;default:current_timestamp;comment:资源最后修改时间" json:"updatedAt"` // 资源最后修改时间
}

// TableName {{.Web.R.GORMModel}}'s table name
func (*{{.Web.R.GORMModel}}) TableName() string {
	return TableName{{.Web.R.GORMModel}}
}
