package migrations

import (
	"gorm.io/gorm"
)

type Product struct {
	ID          uint    `gorm:"primaryKey"`
	Name        string  `gorm:"size:255;not null"`
	Description string  `gorm:"type:text"`
	Price       float64 `gorm:"not null"`
	Stock       uint    `gorm:"not null;default:0"`
}

type Migration20240601000100CreateProductsTable struct {}

func (m *Migration20240601000100CreateProductsTable) Up(db *gorm.DB) error {
	return db.AutoMigrate(&Product{})
}

func (m *Migration20240601000100CreateProductsTable) Down(db *gorm.DB) error {
	return db.Migrator().DropTable("products")
}

// Export the migration struct for plugin system
// This function name must match the struct name exactly
var Migration20240601000100CreateProductsTable_Exported = &Migration20240601000100CreateProductsTable{}