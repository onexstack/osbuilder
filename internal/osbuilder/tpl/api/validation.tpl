package validation

import (
	"context"

	genericvalidation "github.com/onexstack/onexstack/pkg/validation"

	{{.APIImportPath}}
)

// Validate{{.SingularName}}Rules returns a set of validation rules for {{.SingularLower}}-related requests.
func (v *Validator) Validate{{.SingularName}}Rules() genericvalidation.Rules {
	return genericvalidation.Rules{}
}

// ValidateCreate{{.SingularName}}Request validates the fields of a Create{{.SingularName}}Request.
func (v *Validator) ValidateCreate{{.SingularName}}Request(ctx context.Context, rq *{{.APIAlias}}.Create{{.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{.SingularName}}Rules())
}

// ValidateUpdate{{.SingularName}}Request validates the fields of an Update{{.SingularName}}Request.
func (v *Validator) ValidateUpdate{{.SingularName}}Request(ctx context.Context, rq *{{.APIAlias}}.Update{{.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{.SingularName}}Rules())
}

// ValidateDelete{{.SingularName}}Request validates the fields of a Delete{{.SingularName}}Request.
func (v *Validator) ValidateDelete{{.SingularName}}Request(ctx context.Context, rq *{{.APIAlias}}.Delete{{.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{.SingularName}}Rules())
}

// ValidateGet{{.SingularName}}Request validates the fields of a Get{{.SingularName}}Request.
func (v *Validator) ValidateGet{{.SingularName}}Request(ctx context.Context, rq *{{.APIAlias}}.Get{{.SingularName}}Request) error {
	return genericvalidation.ValidateAllFields(rq, v.Validate{{.SingularName}}Rules())
}

// ValidateList{{.SingularName}}Request validates the fields of a List{{.SingularName}}Request, focusing on selected fields ("Offset" and "Limit").
func (v *Validator) ValidateList{{.SingularName}}Request(ctx context.Context, rq *{{.APIAlias}}.List{{.SingularName}}Request) error {
	return genericvalidation.ValidateSelectedFields(rq, v.Validate{{.SingularName}}Rules(), "Offset", "Limit")
}
