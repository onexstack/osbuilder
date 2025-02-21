package handler

import (
	"github.com/gin-gonic/gin"

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/biz"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/validation"
)

// Handler implements a gRPC service.
type Handler struct {
	biz biz.IBiz
	val *validation.Validator
	mws []gin.HandlerFunc
}

type Registrar func(v1 *gin.RouterGroup, h *Handler)

var registrars []Registrar

// NewHandler creates a new instance of Handler.
func NewHandler(biz biz.IBiz, val *validation.Validator, mws ...gin.HandlerFunc) *Handler {
	return &Handler{biz: biz, val: val, mws: mws}
}

func Register(r Registrar) {
	registrars = append(registrars, r)
}

func (h *Handler) InstallAll(v1 *gin.RouterGroup) {
	for _, r := range registrars {
		r(v1, h)
	}
}
