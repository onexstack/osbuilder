package model

import (
	"github.com/onexstack/onexstack/pkg/authn"
	"gorm.io/gorm"

	"{{.D.ModuleName}}/internal/pkg/rid"
)

// AfterCreate generates a postID after creating a database record.
func (m *PostM) AfterCreate(tx *gorm.DB) error {
	m.PostID = rid.PostID.New(uint64(m.ID))

	return tx.Save(m).Error
}

// BeforeCreate encrypts the plaintext password before creating a database record.
func (m *UserM) BeforeCreate(tx *gorm.DB) error {
	// Encrypt the user password.
	var err error
	m.Password, err = authn.Encrypt(m.Password)
	if err != nil {
		return err
	}

	return nil
}

// AfterCreate generates a userID after creating a database record.
func (m *UserM) AfterCreate(tx *gorm.DB) error {
	m.UserID = rid.UserID.New(uint64(m.ID))

	return tx.Save(m).Error
}
