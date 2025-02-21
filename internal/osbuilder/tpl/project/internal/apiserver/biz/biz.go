package biz

import (
	"github.com/google/wire"
	{{- if .Web.WithUser }}
	"github.com/onexstack/onexstack/pkg/authz"
	{{- end}}

	{{- if .Web.WithUser }}
    userv1 "{{.D.ModuleName}}/internal/{{.Web.Name}}/biz/v1/user"
	{{- end}}
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
)

// ProviderSet is a Wire provider set used to declare dependency injection rules.
// Includes the NewBiz constructor to create a biz instance.
// wire.Bind binds the IBiz interface to the concrete implementation *biz,
// so places that depend on IBiz will automatically inject a *biz instance.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz defines the methods that must be implemented by the business layer.
type IBiz interface {
	{{- if .Web.WithUser }}
    // UserV1 获取用户业务接口.
    UserV1() userv1.UserBiz
	{{- end}}

}

// biz is a concrete implementation of IBiz.
type biz struct {
	store store.IStore
	{{- if .Web.WithUser }}
	authz *authz.Authz
	{{- end}}
}

// Ensure that biz implements the IBiz.
var _ IBiz = (*biz)(nil)

// NewBiz creates an instance of IBiz.
func NewBiz(store store.IStore{{- if .Web.WithUser }}, authz *authz.Authz{{- end -}}) *biz {
	return &biz{store: store{{- if .Web.WithUser }}, authz: authz{{end}}}
}

{{- if .Web.WithUser }}
// UserV1 返回一个实现了 UserBiz 接口的实例.
func (b *biz) UserV1() userv1.UserBiz {
    return userv1.New(b.store, b.authz)
}
{{- end}}
