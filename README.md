# Go-Migration

Package Go-Migration adalah package sederhana untuk mengelola migrasi database di aplikasi Go menggunakan GORM.

## Fitur

- Membuat file migrasi dengan timestamp
- Menjalankan migrasi yang belum dijalankan
- Rollback migrasi yang sudah dijalankan
- Pelacakan migrasi yang sudah dijalankan di database

## Instalasi

```bash
go get github.com/tensuqiuwulu/go-migration
```

## Penggunaan

### 1. Konfigurasi Database

Ada dua cara untuk mengatur koneksi database:

#### Cara 1: Menggunakan SetDatabaseConfig

```go
import "github.com/tensuqiuwulu/go-migration/migration"

func main() {
    // Konfigurasi koneksi database
    migration.SetDatabaseConfig("mysql", "user:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local")
    
    // ...
}
```

#### Cara 2: Menginjeksi Koneksi Database yang Sudah Ada

```go
import (
    "github.com/tensuqiuwulu/go-migration/migration"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func main() {
    // Membuat koneksi database di aplikasi client
    dsn := "user:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect to database")
    }
    
    // Menginjeksi koneksi database ke go-migration
    migration.SetDatabaseConnection(db)
    
    // ...
}
```

### 2. Menjalankan Perintah Migrasi

Package ini menyediakan beberapa perintah untuk mengelola migrasi:

```go
import (
    "os"
    "github.com/tensuqiuwulu/go-migration/migration"
)

func main() {
    // Cek argumen command line
    if len(os.Args) > 1 {
        migration.ExecuteCommand(os.Args[1:])
        return
    }
    
    // Jalankan aplikasi utama jika bukan perintah migrasi
    // ...
}
```

### 3. Perintah yang Tersedia

#### Membuat File Migrasi Baru

```bash
go run main.go make:migration nama_migrasi
```

Perintah ini akan membuat file migrasi baru di direktori `migrations/` dengan format `TIMESTAMP_nama_migrasi.go`.

#### Menjalankan Migrasi

```bash
go run main.go migrate
```

Perintah ini akan menjalankan semua migrasi yang belum dijalankan.

#### Rollback Migrasi

```bash
go run main.go migrate:rollback
```

Perintah ini akan melakukan rollback migrasi dari batch terakhir.

## Contoh Implementasi

### Contoh 1: Membuat Tabel dengan SQL

```go
package main

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
```

### Contoh 2: Menggunakan GORM Model

```go
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
```

### Contoh 3: Menambahkan Kolom ke Tabel yang Sudah Ada

```go
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
```

## Integrasi dengan Aplikasi

Berikut adalah contoh cara mengintegrasikan package Go-Migration ke dalam aplikasi Go:

```go
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
		migration.SetDatabaseConfig("mysql", "user:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local")
		
		// Jalankan perintah migration
		migration.ExecuteCommand(os.Args[1:])
		return
	}

	// Jalankan aplikasi utama jika bukan perintah migration
	fmt.Println("Starting application...")
	// ...
}
```

## Catatan Penting

1. Pastikan direktori `migrations/` sudah ada di root project Anda.
2. File migrasi harus mengikuti format yang ditentukan dengan interface `Migration`.
3. Nama struct migrasi harus unik untuk menghindari konflik.
4. Migrasi dijalankan berdasarkan urutan timestamp pada nama file.

## Lisensi

MIT