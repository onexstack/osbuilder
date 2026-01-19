package gin

import (
    "context"
    "log/slog"

    "github.com/gin-gonic/gin"
    "github.com/onexstack/onexstack/pkg/core"
    "github.com/onexstack/onexstack/pkg/token"

    "{{.D.ModuleName}}/internal/pkg/contextx"
    "{{.D.ModuleName}}/internal/pkg/errno"
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/model"
)

// UserRetriever is an interface for retrieving user information based on username.
type UserRetriever interface {
    // GetUser retrieves user information by user ID.
    GetUser(ctx context.Context, userID string) (*model.UserM, error)
}

// AuthnMiddleware is an authentication middleware used to extract and validate the token from gin.Context.
func AuthnMiddleware(retriever UserRetriever) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Parse the JWT Token.
        userID, err := token.ParseRequest(c)
        if err != nil {
            core.WriteResponse(c, nil, errno.ErrTokenInvalid.WithMessage("%s", err.Error()))
            c.Abort()
            return
        }

        slog.Info("Token parsing successful", "userID", userID)

        user, err := retriever.GetUser(c, userID)
        if err != nil {
            core.WriteResponse(c, nil, errno.ErrUnauthenticated.WithMessage("%s", err.Error()))
            c.Abort()
            return
        }

        ctx := contextx.WithUserID(c.Request.Context(), user.UserID)
        ctx = contextx.WithUsername(ctx, user.Username)
        c.Request = c.Request.WithContext(ctx)

        c.Next()
    }
}
