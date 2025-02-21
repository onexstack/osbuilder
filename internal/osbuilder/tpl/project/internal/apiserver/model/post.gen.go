package model

import (
	"time"
)

const TableName{{.Web.R.GORMModel}} = "{{.Web.R.SingularLower}}"

// {{.Web.R.GORMModel}} mapped from table <{{.Web.R.SingularLower}}>
type {{.Web.R.GORMModel}} struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	{{.Web.R.SingularName}}ID    string    `gorm:"column:{{.Web.R.SingularLowerFirst}}ID;not null;comment:资源唯一 ID" json:"{{.Web.R.SingularLowerFirst}}ID"` // 资源唯一 ID
	CreatedAt time.Time `gorm:"column:createdAt;not null;default:current_timestamp;comment:资源创建时间" json:"createdAt"`   // 资源创建时间
	UpdatedAt time.Time `gorm:"column:updatedAt;not null;default:current_timestamp;comment:资源最后修改时间" json:"updatedAt"` // 资源最后修改时间
}

// TableName {{.Web.R.GORMModel}}'s table name
func (*{{.Web.R.GORMModel}}) TableName() string {
	return TableName{{.Web.R.GORMModel}}
}
