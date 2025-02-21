// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

//go:build wireinject
// +build wireinject

package {{.Web.Name}}

import (
	"github.com/google/wire"
	"github.com/onexstack/onexstack/pkg/server"
    {{- if .Web.WithUser}}
    "github.com/onexstack/onexstack/pkg/authz"
    {{- end}}

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/biz"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/validation"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
    {{- if .Web.WithUser}}
    {{- if eq .Web.WebFramework "gin" }}
    mw "{{.D.ModuleName}}/internal/pkg/middleware/gin"
    {{- else if eq .Web.WebFramework "grpc"}}
    mw "{{.D.ModuleName}}/internal/pkg/middleware/grpc"
    {{- end}}
    {{- end}}

)

// InitializeWebServer sets up and initializes the web server with all necessary dependencies.
func InitializeWebServer(*Config) (server.Server, error) {
    wire.Build(
		NewWebServer,
        wire.Struct(new(ServerConfig), "*"), // * 表示注入全部字段
        wire.NewSet(store.ProviderSet, biz.ProviderSet),
        ProvideDB, // 提供数据库实例
        validation.ProviderSet,
        {{- if .Web.WithUser}}
        wire.NewSet(
            wire.Struct(new(UserRetriever), "*"),
            wire.Bind(new(mw.UserRetriever), new(*UserRetriever)),
        ),
        authz.ProviderSet,
        {{- end}}
    )
    return nil, nil
}
