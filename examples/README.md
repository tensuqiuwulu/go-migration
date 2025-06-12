# Contoh Penggunaan Go-Migration

Direktori ini berisi contoh penggunaan package Go-Migration dengan implementasi loadMigrations menggunakan plugin Go.

## Struktur Direktori

```
./
├── go.mod              # File konfigurasi modul Go
├── main.go             # File utama untuk menjalankan migrasi
└── migrations/         # Direktori berisi file-file migrasi
    ├── 20240601000000_create_users_table.go     # Migrasi untuk membuat tabel users
    ├── 20240601000100_create_products_table.go  # Migrasi untuk membuat tabel products
    └── 20240601000200_add_phone_to_users.go     # Migrasi untuk menambahkan kolom phone ke tabel users
```

## Cara Menjalankan

1. Pastikan Anda memiliki database MySQL yang berjalan
2. Sesuaikan konfigurasi database di `main.go`
   - Anda dapat menggunakan `SetDatabaseConfig` (cara 1, sudah aktif secara default)
   - Atau menggunakan `SetDatabaseConnection` (cara 2, perlu uncomment kode di `main.go`)
3. Jalankan perintah berikut:
### Implementasi loadMigrations dengan Plugin Go

Contoh ini menggunakan implementasi `loadMigrations()` yang memanfaatkan fitur plugin Go untuk memuat migrasi secara dinamis. Berikut adalah langkah-langkah yang dilakukan:

1. File migrasi di direktori `migrations/` dikompilasi menjadi plugin Go (file `.so`)
2. Plugin tersebut dimuat secara dinamis saat runtime
3. Struct migrasi dicari berdasarkan nama yang diekstrak dari nama file
4. Migrasi dijalankan sesuai urutan timestamp pada nama file

### Catatan Penting

- Pastikan semua file migrasi menggunakan package `migrations` (bukan `main`)
- Nama struct migrasi harus mengikuti format: `Migration<timestamp><CamelCaseName>`
- Contoh: `Migration20240601000000CreateUsersTable`
- File migrasi harus mengimplementasikan interface `Migration` dengan method `Up()` dan `Down()`
- **Penting**: Setiap file migrasi harus mengekspor struct migrasi sebagai variabel dengan akhiran `_Exported`, contoh:

```go
// Definisi struct migrasi
type Migration20240601000000CreateUsersTable struct {}

// Implementasi interface Migration
func (m *Migration20240601000000CreateUsersTable) Up(db *gorm.DB) error {
    // Implementasi Up
}

func (m *Migration20240601000000CreateUsersTable) Down(db *gorm.DB) error {
    // Implementasi Down
}

// Ekspor struct migrasi untuk sistem plugin
// Nama variabel harus sama persis dengan nama struct + "_Exported"
var Migration20240601000000CreateUsersTable_Exported = &Migration20240601000000CreateUsersTable{}
```

## Cara Menjalankan
### Membuat Migrasi Baru

```bash
go run main.go make:migration nama_migrasi
```

### Menjalankan Migrasi

```bash
go run main.go migrate
```

### Rollback Migrasi

```bash
go run main.go migrate:rollback
```

## Contoh File Migrasi

### 1. Membuat Tabel dengan SQL (create_users_table.go)

Contoh ini menunjukkan cara membuat tabel menggunakan SQL langsung.

### 2. Membuat Tabel dengan GORM Model (create_products_table.go)

Contoh ini menunjukkan cara membuat tabel menggunakan GORM Model dan AutoMigrate.

### 3. Menambahkan Kolom ke Tabel (add_phone_to_users.go)

Contoh ini menunjukkan cara menambahkan kolom baru ke tabel yang sudah ada.