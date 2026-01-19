package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

// Err{{.Web.R.SingularName}}NotFound indicates that the specified {{.Web.R.SingularLower}} was not found.
var Err{{.Web.R.SingularName}}NotFound = errorsx.New(http.StatusNotFound, "NotFound.{{.Web.R.SingularName}}NotFound", "The requested {{.Web.R.SingularLower}} was not found.")

// Err{{.Web.R.SingularName}}CreateFailed indicates that the {{.Web.R.SingularLower}} creation operation failed.
var Err{{.Web.R.SingularName}}CreateFailed = errorsx.New(http.StatusInternalServerError, "InternalError.{{.Web.R.SingularName}}CreateFailed", "Failed to create the {{.Web.R.SingularLower}}.")

// Err{{.Web.R.SingularName}}UpdateFailed indicates that the {{.Web.R.SingularLower}} update operation failed.
var Err{{.Web.R.SingularName}}UpdateFailed = errorsx.New(http.StatusInternalServerError, "InternalError.{{.Web.R.SingularName}}UpdateFailed", "Failed to update the {{.Web.R.SingularLower}}.")

// Err{{.Web.R.SingularName}}DeleteFailed indicates that the {{.Web.R.SingularLower}} deletion operation failed.
var Err{{.Web.R.SingularName}}DeleteFailed = errorsx.New(http.StatusInternalServerError, "InternalError.{{.Web.R.SingularName}}DeleteFailed", "Failed to delete the {{.Web.R.SingularLower}}.")

// Err{{.Web.R.SingularName}}GetFailed indicates that retrieving the specified {{.Web.R.SingularLower}} failed.
var Err{{.Web.R.SingularName}}GetFailed = errorsx.New(http.StatusInternalServerError, "InternalError.{{.Web.R.SingularName}}GetFailed", "Failed to retrieve the {{.Web.R.SingularLower}} details.")

// Err{{.Web.R.SingularName}}ListFailed indicates that listing {{.Web.R.PluralLower}} failed.
var Err{{.Web.R.SingularName}}ListFailed = errorsx.New(http.StatusInternalServerError, "InternalError.{{.Web.R.SingularName}}ListFailed", "Failed to list {{.Web.R.PluralLower}}.")
