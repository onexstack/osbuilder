package {{.APIVersion}}

// Default sets default values for the Create{{.SingularName}}Request struct.
func (rq *Create{{.SingularName}}Request) Default() {}

// Default sets default values for the Update{{.SingularName}}Request struct.
func (rq *Update{{.SingularName}}Request) Default() {}

// Default sets default values for the Delete{{.SingularName}}Request struct.
func (rq *Delete{{.SingularName}}Request) Default() {}

// Default sets default values for the Get{{.SingularName}}Request struct.
func (rq *Get{{.SingularName}}Request) Default() {}

// Default sets default values for the List{{.SingularName}}Request struct.
func (rq *List{{.SingularName}}Request) Default() {}
