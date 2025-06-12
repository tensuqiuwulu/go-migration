package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gorm.io/gorm"
)

// MigrationRecord represents a record in the migrations table
type MigrationRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Migration string    `gorm:"size:255;not null;unique"`
	Batch     int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

type Migration interface {
	Up(*gorm.DB) error
	Down(*gorm.DB) error
}

// CreateMigration membuat file migration baru
func CreateMigration(name string) error {
	// Membuat direktori migrations jika belum ada
	if err := os.MkdirAll("migrations", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Generate timestamp untuk nama file
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.go", timestamp, snakeCase(name))
	filePath := filepath.Join("migrations", filename)

	// Membuat file migration
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}
	defer file.Close()

	// Generate nama struct untuk migration
	structName := fmt.Sprintf("Migration%s%s", timestamp, camelCase(name))

	// Eksekusi template
	tmpl, err := template.New("migration").Parse(migrationTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse migration template: %w", err)
	}

	data := struct {
		StructName string
	}{
		StructName: structName,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to generate migration content: %w", err)
	}

	fmt.Printf("Created new migration: %s\n", filePath)
	return nil
}

// Helper functions
func snakeCase(s string) string {
	// Implementasi konversi ke snake_case
	return strings.ToLower(strings.Join(strings.Fields(s), "_"))
}

func camelCase(s string) string {
	// Implementasi konversi ke CamelCase
	words := strings.Fields(strings.ReplaceAll(s, "_", " "))
	for i := range words {
		if i > 0 {
			words[i] = strings.Title(words[i])
		}
	}
	return strings.Join(words, "")
}
// ensureMigrationsTable ensures that the migrations table exists
func ensureMigrationsTable(db *gorm.DB) error {
	return db.AutoMigrate(&MigrationRecord{})
}

// getMigrationBatch gets the current batch number
func getMigrationBatch(db *gorm.DB) (int, error) {
	var batch int
	result := db.Model(&MigrationRecord{}).Select("COALESCE(MAX(batch), 0) + 1").Scan(&batch)
	if result.Error != nil {
		return 0, result.Error
	}
	return batch, nil
}

// getMigratedNames gets the names of migrations that have already been run
func getMigratedNames(db *gorm.DB) (map[string]bool, error) {
	var records []MigrationRecord
	result := db.Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}

	migratedNames := make(map[string]bool)
	for _, record := range records {
		migratedNames[record.Migration] = true
	}

	return migratedNames, nil
}

// recordMigration records that a migration has been run
func recordMigration(db *gorm.DB, name string, batch int) error {
	return db.Create(&MigrationRecord{
		Migration: name,
		Batch:     batch,
		CreatedAt: time.Now(),
	}).Error
}

// getLastBatchMigrations gets the migrations from the last batch
func getLastBatchMigrations(db *gorm.DB) ([]string, error) {
	var lastBatch int
	result := db.Model(&MigrationRecord{}).Select("MAX(batch)").Scan(&lastBatch)
	if result.Error != nil {
		return nil, result.Error
	}

	var records []MigrationRecord
	result = db.Where("batch = ?", lastBatch).Order("id DESC").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}

	migrations := make([]string, len(records))
	for i, record := range records {
		migrations[i] = record.Migration
	}

	return migrations, nil
}

// removeMigrationRecord removes a migration record
func removeMigrationRecord(db *gorm.DB, name string) error {
	return db.Where("migration = ?", name).Delete(&MigrationRecord{}).Error
}

func RunMigrations(db *gorm.DB, migrations []Migration) error {
	// Ensure migrations table exists
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current batch number
	batch, err := getMigrationBatch(db)
	if err != nil {
		return fmt.Errorf("failed to get migration batch: %w", err)
	}

	// Get already migrated names
	migratedNames, err := getMigratedNames(db)
	if err != nil {
		return fmt.Errorf("failed to get migrated names: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		// Get migration name from type
		migrationName := fmt.Sprintf("%T", migration)

		// Skip if already migrated
		if migratedNames[migrationName] {
			fmt.Printf("Skipping migration %s (already run)\n", migrationName)
			continue
		}

		fmt.Printf("Running migration %s...\n", migrationName)

		// Run migration
		if err := migration.Up(db); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migrationName, err)
		}

		// Record migration
		if err := recordMigration(db, migrationName, batch); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
		}

		fmt.Printf("Migration %s completed\n", migrationName)
	}

	return nil
}

func RollbackMigrations(db *gorm.DB, migrations []Migration) error {
	// Ensure migrations table exists
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get migrations from last batch
	lastBatchMigrations, err := getLastBatchMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get last batch migrations: %w", err)
	}

	if len(lastBatchMigrations) == 0 {
		fmt.Println("Nothing to rollback")
		return nil
	}

	// Create a map for quick lookup
	migrationMap := make(map[string]Migration)
	for _, migration := range migrations {
		migrationMap[fmt.Sprintf("%T", migration)] = migration
	}

	// Rollback migrations in reverse order
	for _, migrationName := range lastBatchMigrations {
		migration, ok := migrationMap[migrationName]
		if !ok {
			return fmt.Errorf("migration %s not found", migrationName)
		}

		fmt.Printf("Rolling back migration %s...\n", migrationName)

		// Run down migration
		if err := migration.Down(db); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migrationName, err)
		}

		// Remove migration record
		if err := removeMigrationRecord(db, migrationName); err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", migrationName, err)
		}

		fmt.Printf("Rolled back migration %s\n", migrationName)
	}

	return nil
}