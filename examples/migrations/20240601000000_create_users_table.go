package migrations

import (
	"gorm.io/gorm"
)

type Migration20240601000000CreateUsersTable struct {}

func (m *Migration20240601000000CreateUsersTable) Up(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`).Error
}

func (m *Migration20240601000000CreateUsersTable) Down(db *gorm.DB) error {
	return db.Exec(`DROP TABLE IF EXISTS users`).Error
}

// Export the migration struct for plugin system
// This function name must match the struct name exactly
var Migration20240601000000CreateUsersTable_Exported = &Migration20240601000000CreateUsersTable{}