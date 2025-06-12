package main

import (
	"fmt"
	"os"

	"github.com/tensuqiuwulu/go-migration/migration"
	// Uncomment baris berikut untuk menggunakan koneksi database langsung
	// "gorm.io/driver/mysql"
	// "gorm.io/gorm"
	
	// Plugin package diperlukan untuk loadMigrations
	_ "plugin"
)

func main() {
	// Cek apakah ada argumen untuk migration
	if len(os.Args) > 1 && (os.Args[1] == "make:migration" || os.Args[1] == "migrate" || os.Args[1] == "migrate:rollback") {
		// Cara 1: Menggunakan SetDatabaseConfig
		migration.SetDatabaseConfig("mysql", "root:password@tcp(localhost:3306)/example_db?charset=utf8mb4&parseTime=True&loc=Local")
		
		// Cara 2: Menginjeksi koneksi database yang sudah ada
		// Uncomment kode berikut untuk menggunakan cara 2
		/*
		dsn := "root:password@tcp(localhost:3306)/example_db?charset=utf8mb4&parseTime=True&loc=Local"
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			panic("failed to connect to database")
		}
		migration.SetDatabaseConnection(db)
		*/
		
		// CATATAN PENTING:
		// Implementasi loadMigrations() menggunakan plugin Go untuk memuat migrasi secara dinamis.
		// Pastikan semua file migrasi di direktori migrations/ menggunakan package "migrations".
		// Struktur migrasi harus mengikuti format: Migration<timestamp><CamelCaseName>
		// Contoh: Migration20240601000000CreateUsersTable
		
		// Jalankan perintah migration
		migration.ExecuteCommand(os.Args[1:])
		return
	}

	// Tampilkan bantuan jika tidak ada argumen
	fmt.Println("Go-Migration Example")
	fmt.Println("Available commands:")
	fmt.Println("  make:migration <name> - Create a new migration file")
	fmt.Println("  migrate - Run all pending migrations")
	fmt.Println("  migrate:rollback - Rollback the last batch of migrations")
}