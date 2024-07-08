package migrator

import (
	"errors"
	"flag"
	"fmt"
)
import "github.com/golang-migrate/migrate/v4"

func main() {
	var storagePath, migratorPath, migrationTable string

	flag.StringVar(&storagePath, "storage-path", "", "path to storage")
	flag.StringVar(&migratorPath, "migrator-path", "", "path to migrator")
	flag.StringVar(&migrationTable, "migration-table", "", "migration table name")
	flag.Parse()

	if storagePath == "" {
		panic("storage path is required")
	}

	if migratorPath == "" {
		panic("migrator path is required")
	}

	m, err := migrate.New(
		"file://"+migratorPath,
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationTable),
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return
		}

		panic(err)
	}

	fmt.Println("migrations applied")
}
