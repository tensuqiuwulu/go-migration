package main

import (
	"fmt"
	"os"

	"github.com/tensuqiuwulu/go-migration/migration"
)

func main() {
	// Cek apakah ada argumen untuk migration
	if len(os.Args) > 1 && (os.Args[1] == "make:migration" || os.Args[1] == "migrate" || os.Args[1] == "migrate:rollback") {
		// Konfigurasi database
		migration.SetDatabaseConfig("mysql", "root:password@tcp(localhost:3306)/example_db?charset=utf8mb4&parseTime=True&loc=Local")
		
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