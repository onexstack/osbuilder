package gin

import (
    "log/slog"

    "github.com/gin-gonic/gin"
    "github.com/onexstack/onexstack/pkg/core"

    "{{.D.ModuleName}}/internal/pkg/contextx"
    "{{.D.ModuleName}}/internal/pkg/errno"
)

// Authorizer defines the interface for authorization implementation.
type Authorizer interface {
    Authorize(subject, object, action string) (bool, error)
}

// AuthzMiddleware is a Gin middleware for request authorization.
func AuthzMiddleware(authorizer Authorizer) gin.HandlerFunc {
    return func(c *gin.Context) {
        subject := contextx.UserID(c.Request.Context())
        object := c.Request.URL.Path
        action := c.Request.Method

        // Log authorization context information.
        slog.Info("Build authorize context", "subject", subject, "object", object, "action", action)

        // Call the authorization interface for verification.
        if allowed, err := authorizer.Authorize(subject, object, action); err != nil || !allowed {
            core.WriteResponse(c, nil, errno.ErrPermissionDenied.WithMessage(
                "access denied: subject=%s, object=%s, action=%s, reason=%v",
                subject,
                object,
                action,
                err,
            ))
            c.Abort()
            return
        }

        c.Next() // Continue processing the request.
    }
}
