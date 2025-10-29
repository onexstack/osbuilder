package model

import (
	"gorm.io/gorm"
	"github.com/onexstack/onexstack/pkg/store/registry"
	"github.com/onexstack/onexstack/pkg/rid"
)

// AfterCreate generates a {{.Web.R.SingularLowerFirst}}ID after creating a database record.
func (m *{{.Web.R.GORMModel}}) AfterCreate(tx *gorm.DB) error {
	m.{{.Web.R.SingularName}}ID = rid.NewResourceID("{{.Web.R.SingularLower}}").New(uint64(m.ID))

	return tx.Save(m).Error
}

func init() {
	registry.Register(&{{.Web.R.GORMModel}}{})
}
