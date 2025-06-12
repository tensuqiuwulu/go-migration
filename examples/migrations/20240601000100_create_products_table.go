package main

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