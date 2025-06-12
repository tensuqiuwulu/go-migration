package main

import (
	"gorm.io/gorm"
)

type Migration20240601000200AddPhoneToUsers struct {}

func (m *Migration20240601000200AddPhoneToUsers) Up(db *gorm.DB) error {
	return db.Exec(`ALTER TABLE users ADD COLUMN phone VARCHAR(20) NULL AFTER email`).Error
}

func (m *Migration20240601000200AddPhoneToUsers) Down(db *gorm.DB) error {
	return db.Exec(`ALTER TABLE users DROP COLUMN phone`).Error
}