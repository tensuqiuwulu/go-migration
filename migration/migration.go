package migration

const migrationTemplate = `package migrations

import (
	"gorm.io/gorm"
)

type {{.StructName}} struct {}

func (m *{{.StructName}}) Up(db *gorm.DB) error {
	// Implement your migration here
	return nil
}

func (m *{{.StructName}}) Down(db *gorm.DB) error {
	// Implement your rollback here
	return nil
}
`