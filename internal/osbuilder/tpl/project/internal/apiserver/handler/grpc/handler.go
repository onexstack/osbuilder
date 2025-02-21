package handler

import (
	"github.com/google/wire"

    "{{.D.ModuleName}}/internal/{{.Web.Name}}/biz"
	{{.Web.APIImportPath}}
)

// ProviderSet contains providers for creating instances of the biz struct.
var ProviderSet = wire.NewSet(NewHandler, wire.Bind(new({{.D.APIVersion}}.{{.Web.GRPCServiceName}}Server), new(*Handler)))

// Handler implements a gRPC service.
type Handler struct {
	{{.D.APIAlias}}.Unimplemented{{.Web.GRPCServiceName}}Server

	biz biz.IBiz
}

// Ensure that Handler implements the {{.D.APIVersion}}.{{.Web.GRPCServiceName}}Server interface.
var _ {{.D.APIVersion}}.{{.Web.GRPCServiceName}}Server = (*Handler)(nil)

// NewHandler creates a new instance of *Handler.
func NewHandler(biz biz.IBiz) *Handler {
	return &Handler{biz: biz}
}
