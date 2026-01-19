package handler

import (
    "log/slog"

    "github.com/gin-gonic/gin"
    "github.com/onexstack/onexstack/pkg/core"
    {{- if .Web.WithOTel}}
    "go.opentelemetry.io/otel"

    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/metrics"
    {{- end}}
)

// Login logs in a user and returns a JWT Token.
func (h *Handler) Login(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().Login, h.val.ValidateLoginRequest)
}

// RefreshToken refreshes the JWT Token.
func (h *Handler) RefreshToken(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().RefreshToken)
}

// ChangePassword changes the user's password.
func (h *Handler) ChangePassword(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.UserV1().ChangePassword, h.val.ValidateChangePasswordRequest)
}

// CreateUser handles the HTTP request to create a new user.
func (h *Handler) CreateUser(c *gin.Context) {
	{{- if .Web.WithOTel}}
    ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "Handler.CreateUser")
    defer span.End()

    // Update the Gin request context so subsequent middleware/handlers use the traced context.
    c.Request = c.Request.WithContext(ctx)

    metrics.M.RecordResourceCreate(ctx, "user")
	{{- end}}

    slog.InfoContext(ctx, "processing user creation request")

    core.HandleJSONRequest(c, h.biz.UserV1().Create, h.val.ValidateCreateUserRequest)
}

// UpdateUser handles the HTTP request to update an existing user's details.
func (h *Handler) UpdateUser(c *gin.Context) {
	core.HandleAllRequest(c, h.biz.UserV1().Update, h.val.ValidateUpdateUserRequest)
}

// DeleteUser handles the HTTP request to delete a single user specified by URI parameters.
func (h *Handler) DeleteUser(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.UserV1().Delete, h.val.ValidateDeleteUserRequest)
}

// GetUser retrieves details of a specific user based on the request parameters.
func (h *Handler) GetUser(c *gin.Context) {
	{{- if .Web.WithOTel}}
    ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "Handler.GetUser")
    defer span.End()

    c.Request = c.Request.WithContext(ctx)

    metrics.M.RecordResourceGet(ctx, "user")
	{{- end}}

    slog.InfoContext(ctx, "processing user retrieve request", "layer", "handler")

    core.HandleUriRequest(c, h.biz.UserV1().Get, h.val.ValidateGetUserRequest)
}

// ListUser retrieves a list of users based on query parameters.
func (h *Handler) ListUser(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.UserV1().List, h.val.ValidateListUserRequest)
}

func init() {
	Register(func(v1 *gin.RouterGroup, handler *Handler, mws ...gin.HandlerFunc) {
		rg := v1.Group("/users")
		rg.POST("", handler.CreateUser)
		rg.Use(mws...)
		rg.PUT(":userID/change-password", handler.ChangePassword)
		rg.PUT(":userID", handler.UpdateUser)
		rg.DELETE(":userID", handler.DeleteUser)
		rg.GET(":userID", handler.GetUser)
		rg.GET("", handler.ListUser)
	})
}
