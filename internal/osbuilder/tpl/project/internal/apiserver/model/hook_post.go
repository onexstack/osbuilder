package model

import (
	"github.com/onexstack/onexstack/pkg/rid"
	"github.com/onexstack/onexstack/pkg/store/registry"
	"gorm.io/gorm"
)

// AfterCreate generates and updates the {{.Web.R.Last.SingularName}}ID after the database record is created.
func (m *{{.Web.R.GORMModel}}) AfterCreate(tx *gorm.DB) error {
	// Generate the resource ID based on the auto-increment primary key.
	m.{{.Web.R.Last.SingularName}}ID = rid.NewResourceID("{{.Web.R.Last.SingularLower}}").New(uint64(m.ID))

	// Update only the {{.Web.R.Last.SingularName}}ID column to avoid overhead and side effects of a full Save.
	// UpdateColumn is faster as it doesn't update timestamps or trigger Update hooks.
	return tx.Model(m).UpdateColumn("{{.Web.R.Last.SingularLower}}_id", m.{{.Web.R.Last.SingularName}}ID).Error
}

func init() {
	registry.Register(&{{.Web.R.GORMModel}}{})
}
