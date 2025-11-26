package clientset

import (
	{{- range .Web.Clients }}
    "{{$.D.ModuleName}}/internal/{{$.Web.Name}}/pkg/clientset/typed/{{. | lowerkind}}"
    {{- end}}

)

// Interface defines the operations for accessing different client types.
type Interface interface {
	{{- range .Web.Clients }}
	{{. | kind}}() {{. | lowerkind}}.Interface
    {{- end}}
}

// Clientset provides access to different typed clients.
type Clientset struct {
	{{- range .Web.Clients }}
	{{. | lowerkind}} {{. | lowerkind}}.Interface
    {{- end}}
}

// New creates a new Clientset with the provided client interfaces.
func New(
	{{- range .Web.Clients }}
	{{. | lowerkind}} {{. | lowerkind}}.Interface,
    {{- end}}
) *Clientset {
	return &Clientset{
		{{- range .Web.Clients }}
		{{. | lowerkind}}: {{. | lowerkind}},
    	{{- end}}
	}
}

{{- range .Web.Clients }}
// {{. | kind}} returns the {{. | lowerkind}} client interface.
func (c *Clientset) {{. | kind}}() {{. | lowerkind}}.Interface {
	return c.{{. | lowerkind}}
}
{{- end}}
