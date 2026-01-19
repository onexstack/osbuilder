package biz

import (
	"github.com/google/wire"
    {{- if .Web.WithUser }}
    "github.com/onexstack/onexstack/pkg/authz"
    {{- end}}

    "{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
    {{- if .Web.WithUser }}
    userv1 "{{.D.ModuleName}}/internal/{{.Web.Name}}/biz/v1/user"
    {{- end}}
    {{- if .Web.Clients }}
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/clientset"
    {{- end}}
    {{- if .Web.WithWS }}
    wsv1 "{{.D.ModuleName}}/internal/{{.Web.Name}}/biz/v1/websocket"
    {{- end}}
)

// ProviderSet declares dependency injection rules for the business logic layer.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz defines the access points for various business logic modules.
type IBiz interface {
    {{- if .Web.WithUser }}
    // UserV1 gets the user business interface.
    UserV1() userv1.UserBiz
    {{- end}}
    {{- if .Web.WithWS }}
    // WSV1 gets the WebSocket related interface.
    WSV1() wsv1.WSBiz
    {{- end}}
}

// biz is the concrete implementation of the business logic IBiz.
type biz struct {
    store store.IStore
    {{- if .Web.WithUser }}
    authz *authz.Authz
    {{- end}}
    {{- if .Web.Clients }}
    clientset clientset.Interface
    {{- end}}
}

// Ensure biz implements IBiz at compile time.
var _ IBiz = (*biz)(nil)

// NewBiz creates and returns a new instance of the business logic layer.
func NewBiz(store store.IStore{{- if .Web.WithUser }}, authz *authz.Authz{{- end -}}{{- if .Web.Clients }}, clientset clientset.Interface{{- end -}}) *biz {
    return &biz{store: store{{- if .Web.WithUser }}, authz: authz{{end}}{{- if .Web.Clients }}, clientset: clientset{{- end -}}}
}

{{- if .Web.WithUser }}
// UserV1 returns an instance that implements the UserBiz interface.
func (b *biz) UserV1() userv1.UserBiz {
    return userv1.New(b.store, b.authz)
}
{{- end}}

{{- if .Web.WithWS }}
// WSV1 returns an instance that implements the WSBiz interface.
func (b *biz) WSV1() wsv1.WSBiz {
    return wsv1.New(b.store{{- if .Web.Clients }}, b.clientset{{- end -}})
}
{{- end}}
