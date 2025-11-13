package validation

import (
	"context"

	genericvalidation "github.com/onexstack/onexstack/pkg/validation"

	{{.Web.APIImportPath}}
)

// Validate{{.Web.R.SingularName}}Rules returns a set of validation rules for {{.Web.R.SingularLower}}-related requests.
func (v *Validator) Validate{{.Web.R.SingularName}}Rules() genericvalidation.Rules {
	return genericvalidation.Rules{}
}

// ValidateCreate{{.Web.R.SingularName}}Request validates the fields of a Create{{.Web.R.SingularName}}Request.
func (v *Validator) ValidateCreate{{.Web.R.SingularName}}Request(ctx context.Context, rq *{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{.Web.R.SingularName}}Rules())
}

// ValidateUpdate{{.Web.R.SingularName}}Request validates the fields of an Update{{.Web.R.SingularName}}Request.
func (v *Validator) ValidateUpdate{{.Web.R.SingularName}}Request(ctx context.Context, rq *{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{.Web.R.SingularName}}Rules())
}

// ValidateDelete{{.Web.R.SingularName}}Request validates the fields of a Delete{{.Web.R.SingularName}}Request.
func (v *Validator) ValidateDelete{{.Web.R.SingularName}}Request(ctx context.Context, rq *{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{.Web.R.SingularName}}Rules())
}

// ValidateDelete{{.Web.R.PluralName}}Request validates the fields of a Delete{{.Web.R.PluralName}}Request.
func (v *Validator) ValidateDelete{{.Web.R.PluralName}}Request(ctx context.Context, rq *v1.Delete{{.Web.R.PluralName}}Request) error {
    return genericvalidation.ValidateAllFields(rq, v.Validate{{.Web.R.SingularName}}Rules())
}

// ValidateGet{{.Web.R.SingularName}}Request validates the fields of a Get{{.Web.R.SingularName}}Request.
func (v *Validator) ValidateGet{{.Web.R.SingularName}}Request(ctx context.Context, rq *{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{.Web.R.SingularName}}Rules())
}

// ValidateList{{.Web.R.SingularName}}Request validates the fields of a List{{.Web.R.SingularName}}Request, focusing on selected fields ("Offset" and "Limit").
func (v *Validator) ValidateList{{.Web.R.SingularName}}Request(ctx context.Context, rq *{{.D.APIAlias}}.List{{.Web.R.SingularName}}Request) error {
	return genericvalidation.ValidateSelectedFields(rq, v.Validate{{.Web.R.SingularName}}Rules(), "Offset", "Limit")
}
