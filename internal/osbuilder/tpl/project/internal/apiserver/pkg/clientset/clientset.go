package clientset

import (
	{{- range .Web.Clients }}
    "{{$.D.ModuleName}}/internal/{{$.Web.Name}}/pkg/clientset/typed/{{. | lowerkind}}"
    {{- end}}
)

// Interface defines the operations for accessing different client types within the clientset.
type Interface interface {
	{{- range .Web.Clients }}
	// {{. | kind}} returns the client interface for managing {{. | lowerkind}} resources.
	{{. | kind}}() {{. | lowerkind}}.Interface
    {{- end}}
}

// Clientset provides a unified entry point to access various typed clients.
type Clientset struct {
	{{- range .Web.Clients }}
	{{. | lowerkind}} {{. | lowerkind}}.Interface
    {{- end}}
}

// New creates a new Clientset with the provided client interfaces.
// It acts as a constructor for the Clientset type.
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
// {{. | kind}} returns the client interface for {{. | lowerkind}} resources.
func (c *Clientset) {{. | kind}}() {{. | lowerkind}}.Interface {
	return c.{{. | lowerkind}}
}
{{- end}}
