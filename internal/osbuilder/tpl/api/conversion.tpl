package conversion

import (
	"github.com/jinzhu/copier"

	"{{.ModuleName}}/internal/{{.ComponentName}}/model"
	{{.APIAlias}} "{{.ModuleName}}/pkg/api/{{.ComponentName}}/{{.APIVersion}}"
)

// {{.MapModelToAPIFunc}} converts a {{.GORMModel}} object from the internal model
// to a {{.SingularName}} object in the {{.APIAlias}} API format.
func {{.MapModelToAPIFunc}}({{.SingularLowerFirst}}Model *model.{{.GORMModel}}) *{{.APIAlias}}.{{.SingularName}} {
	var {{.SingularLowerFirst}} {{.APIAlias}}.{{.SingularName}}
	_ = copier.Copy(&{{.SingularLowerFirst}}, {{.SingularLowerFirst}}Model)
	return &{{.SingularLowerFirst}}
}

// {{.MapAPIToModelFunc}} converts a {{.SingularName}} object from the {{.APIAlias}} API format
// to a {{.GORMModel}} object in the internal model.
func {{.MapAPIToModelFunc}}({{.SingularLowerFirst}} *{{.APIAlias}}.{{.SingularName}}) *model.{{.GORMModel}} {
	var {{.SingularLowerFirst}}Model model.{{.GORMModel}}
	_ = copier.Copy(&{{.SingularLowerFirst}}Model, {{.SingularLowerFirst}})
	return &{{.SingularLowerFirst}}Model
}

