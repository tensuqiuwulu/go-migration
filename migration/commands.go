package migration

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"gorm.io/gorm"
)

// Database connection configuration
var (
	dbDialect string = "mysql"
	dbDSN     string = "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	dbConnection *gorm.DB = nil
)

// SetDatabaseConfig sets the database configuration
func SetDatabaseConfig(dialect, dsn string) {
	dbDialect = dialect
	dbDSN = dsn
}

// SetDatabaseConnection sets an existing database connection
func SetDatabaseConnection(db *gorm.DB) {
	dbConnection = db
}

// getDatabase returns a database connection
func getDatabase() (*gorm.DB, error) {
	// If a database connection has been injected, use it
	if dbConnection != nil {
		return dbConnection, nil
	}
	
	// Otherwise, this is a placeholder. In a real implementation, you would use GORM to connect to the database
	// For example:
	// import "gorm.io/driver/mysql"
	// return gorm.Open(mysql.Open(dbDSN), &gorm.Config{})
	//
	// Or for PostgreSQL:
	// import "gorm.io/driver/postgres"
	// return gorm.Open(postgres.Open(dbDSN), &gorm.Config{})
	
	// For now, we'll return an error to remind users to either:
	// 1. Implement this function with the appropriate database driver
	// 2. Inject a database connection using SetDatabaseConnection
	
	return nil, fmt.Errorf("database connection not implemented, please either:\n" +
		"1. Implement getDatabase() function with your database driver\n" +
		"2. Inject a database connection using SetDatabaseConnection")
}

// loadMigrations loads all migrations from the migrations directory
func loadMigrations() ([]Migration, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	log.Printf("Current working directory: %s", cwd)
	
	// Check if migrations directory exists
	migrationsDir := "migrations"
	migrationsPath := filepath.Join(cwd, migrationsDir)
	
	// Check if migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Printf("Migrations directory not found at: %s", migrationsPath)
		return nil, fmt.Errorf("migrations directory not found: %w", err)
	}
	
	// Get all migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// log files
	log.Printf("Found %d files in migrations directory", len(files))
	for i, file := range files {
		log.Printf("Migration file %d: %s", i+1, file.Name())
	}
	
	// Sort files by name to ensure migrations run in order
	filenames := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".go" {
			filenames = append(filenames, file.Name())
		}
	}
	sort.Strings(filenames)
	
	// Compile the migrations directory into a plugin
	log.Printf("Compiling migrations directory into plugin...")
	
	// Use absolute path for migrations directory and output file
	pluginOutputPath := filepath.Join(cwd, "migrations.so")
	
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", pluginOutputPath, migrationsPath)
	
	// Capture command output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	log.Printf("Running command: go build -buildmode=plugin -o %s %s", pluginOutputPath, migrationsPath)
	
	if cmdErr := cmd.Run(); cmdErr != nil {
		log.Printf("Error compiling migrations: %v", cmdErr)
		return nil, fmt.Errorf("failed to compile migrations: %w", cmdErr)
	}
	
	log.Printf("Successfully compiled migrations plugin at: %s", pluginOutputPath)
	
	// Load the plugin using absolute path
	log.Printf("Loading plugin from: %s", pluginOutputPath)
	p, err := plugin.Open(pluginOutputPath)
	if err != nil {
		log.Printf("Error loading plugin: %v", err)
		return nil, fmt.Errorf("failed to load migrations plugin: %w", err)
	}
	
	log.Printf("Successfully loaded migrations plugin")
	
	// Load each migration
	log.Printf("Loading migrations from plugin...")
	migrations := make([]Migration, 0, len(filenames))
	
	for i, filename := range filenames {
		log.Printf("Processing migration file %d/%d: %s", i+1, len(filenames), filename)
		
		// Extract struct name from filename
		// Format: YYYYMMDDHHMMSS_migration_name.go
		parts := strings.Split(strings.TrimSuffix(filename, ".go"), "_")
		if len(parts) < 2 {
			log.Printf("Skipping %s: doesn't follow naming convention", filename)
			continue // Skip files that don't follow the naming convention
		}
		
		// Construct the struct name: Migration<timestamp><CamelCaseName>
		timestamp := parts[0]
		nameParts := parts[1:]
		
		// Convert to camel case
		var camelCaseName string
		for _, part := range nameParts {
			camelCaseName += cases.Title(language.Und).String(part)
		}
		
		// The struct name in the migration file is already prefixed with "Migration"
		structName := "Migration" + timestamp + camelCaseName
		// The exported variable name has _Exported suffix
		exportedVarName := structName + "_Exported"
		log.Printf("Looking for exported variable: %s", exportedVarName)
		
		// Look up exported variable in the plugin
		sym, err := p.Lookup(exportedVarName)
		if err != nil {
			log.Printf("Error looking up exported variable %s: %v", exportedVarName, err)
			// Try the original struct name as fallback
			sym, err = p.Lookup(structName)
			if err != nil {
				log.Printf("Error looking up migration %s: %v", structName, err)
				return nil, fmt.Errorf("failed to lookup migration %s or %s: %w", exportedVarName, structName, err)
			}
			log.Printf("Found migration using original struct name: %s", structName)
		} else {
			log.Printf("Found migration using exported variable: %s", exportedVarName)
		}
		
		log.Printf("Found symbol for %s, checking if it implements Migration interface", structName)
		
		// Try to convert the symbol to a Migration interface
		log.Printf("Symbol type: %T", sym)
		
		// First, check if it's a pointer to a struct that implements Migration
		if migration, ok := sym.(Migration); ok {
			log.Printf("Successfully loaded migration: %s (direct interface)", structName)
			migrations = append(migrations, migration)
			continue
		}
		
		// Next, check if it's a pointer to a struct that we need to dereference
		// This is a more generic approach that doesn't rely on knowing the exact struct type
		valueOfSym := reflect.ValueOf(sym)
		log.Printf("Symbol value kind: %v, can interface? %v", valueOfSym.Kind(), valueOfSym.CanInterface())
		
		if valueOfSym.Kind() == reflect.Ptr && valueOfSym.Elem().CanInterface() {
			// Try to get the concrete value and check if it implements Migration
			concrete := valueOfSym.Elem().Interface()
			log.Printf("Concrete type: %T", concrete)
			
			if migration, ok := concrete.(Migration); ok {
				log.Printf("Successfully loaded migration: %s (via reflection)", structName)
				migrations = append(migrations, migration)
				continue
			}
		}
		
		// If we get here, we couldn't convert the symbol to a Migration
		log.Printf("%s does not implement Migration interface", structName)
		return nil, fmt.Errorf("%s does not implement Migration interface", structName)
	}
	
	log.Printf("Successfully loaded %d migrations", len(migrations))
	
	return migrations, nil
}

func ExecuteCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Available commands:")
		fmt.Println("  make:migration <name> - Create a new migration file")
		fmt.Println("  migrate - Run all pending migrations")
		fmt.Println("  migrate:rollback - Rollback the last batch of migrations")
		return
	}

	switch args[0] {
	case "make:migration":
		if len(args) < 2 {
			fmt.Println("Please specify migration name")
			return
		}
		if err := CreateMigration(args[1]); err != nil {
			fmt.Printf("Error creating migration: %v\n", err)
		} else {
			fmt.Println("Migration created successfully")
		}
	case "migrate":
		fmt.Println("Running migrations...")
		db, err := getDatabase()
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			return
		}
		
		migrations, err := loadMigrations()
		if err != nil {
			fmt.Printf("Error loading migrations: %v\n", err)
			return
		}
		
		if err := RunMigrations(db, migrations); err != nil {
			fmt.Printf("Error running migrations: %v\n", err)
		} else {
			fmt.Println("Migrations completed successfully")
		}
		
	case "migrate:rollback":
		fmt.Println("Rolling back migrations...")
		db, err := getDatabase()
		if err != nil {
			fmt.Printf("Error connecting to database: %v\n", err)
			return
		}
		
		migrations, err := loadMigrations()
		if err != nil {
			fmt.Printf("Error loading migrations: %v\n", err)
			return
		}
		
		if err := RollbackMigrations(db, migrations); err != nil {
			fmt.Printf("Error rolling back migrations: %v\n", err)
		} else {
			fmt.Println("Rollback completed successfully")
		}
		
	default:
		fmt.Println("Unknown command")
	}
}