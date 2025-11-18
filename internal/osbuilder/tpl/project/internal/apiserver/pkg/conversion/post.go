package conversion

import (
	"github.com/onexstack/onexstack/pkg/core"

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/model"
	{{.Web.APIImportPath}}
)

// {{.Web.R.MapModelToAPIFunc}} converts a {{.Web.R.GORMModel}} object from the internal model
// to a {{.Web.R.SingularName}} object in the {{.D.APIAlias}} API format.
func {{.Web.R.MapModelToAPIFunc}}({{.Web.R.Last.SingularLowerFirst}}M *model.{{.Web.R.GORMModel}}) *{{.D.APIAlias}}.{{.Web.R.SingularName}} {
	var {{.Web.R.Last.SingularLowerFirst}} {{.D.APIAlias}}.{{.Web.R.SingularName}}
	_ = core.CopyWithConverters(&{{.Web.R.Last.SingularLowerFirst}}, {{.Web.R.Last.SingularLowerFirst}}M)
	return &{{.Web.R.Last.SingularLowerFirst}}
}

// {{.Web.R.MapAPIToModelFunc}} converts a {{.Web.R.SingularName}} object from the {{.D.APIAlias}} API format
// to a {{.Web.R.GORMModel}} object in the internal model.
func {{.Web.R.MapAPIToModelFunc}}({{.Web.R.Last.SingularLowerFirst}} *{{.D.APIAlias}}.{{.Web.R.SingularName}}) *model.{{.Web.R.GORMModel}} {
	var {{.Web.R.Last.SingularLowerFirst}}M model.{{.Web.R.GORMModel}}
	_ = core.CopyWithConverters(&{{.Web.R.Last.SingularLowerFirst}}M, {{.Web.R.Last.SingularLowerFirst}})
	return &{{.Web.R.Last.SingularLowerFirst}}M
}

