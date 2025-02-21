// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

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
