package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

// Err{{.Web.R.SingularName}}NotFound indicates that the specified {{.Web.R.SingularLower}} was not found.
var Err{{.Web.R.SingularName}}NotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.{{.Web.R.SingularName}}NotFound", Message: "{{.Web.R.SingularName}} not found."}

// Err{{.Web.R.SingularName}}CreateFailed indicates that the specified {{.Web.R.SingularLower}} was failed to create.
var Err{{.Web.R.SingularName}}CreateFailed = &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.{{.Web.R.SingularName}}CreateFailed", Message: "{{.Web.R.SingularName}} create failed."}

// Err{{.Web.R.SingularName}}UpdateFailed indicates that the specified {{.Web.R.SingularLower}} was failed to update.
var Err{{.Web.R.SingularName}}UpdateFailed= &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.{{.Web.R.SingularName}}UpdateFailed", Message: "{{.Web.R.SingularName}} update failed."}

// Err{{.Web.R.SingularName}}DeleteFailed indicates that the specified {{.Web.R.SingularLower}} was failed to delete.
var Err{{.Web.R.SingularName}}DeleteFailed= &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.{{.Web.R.SingularName}}DeleteFailed", Message: "{{.Web.R.SingularName}} delete failed."}

// Err{{.Web.R.SingularName}}GetFailed indicates that the specified {{.Web.R.SingularLower}} was failed to retrive.
var Err{{.Web.R.SingularName}}GetFailed= &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.{{.Web.R.SingularName}}GetFailed", Message: "{{.Web.R.SingularName}} retrive failed."}

// Err{{.Web.R.SingularName}}ListFailed indicates that the specified {{.Web.R.SingularLower}} was failed to list.
var Err{{.Web.R.SingularName}}ListFailed= &errorsx.ErrorX{Code: http.StatusInternalServerError, Reason: "InternalError.{{.Web.R.SingularName}}ListFailed", Message: "{{.Web.R.SingularName}} list failed."}
