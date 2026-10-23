package main

import (
	"errors"
	"fmt"

	"flag"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		dbHost          string
		dbPort          string
		dbUser          string
		dbPassword      string
		dbName          string
		sslMode         string
		migrationsPath  string
		migrationsTable string
	)

	flag.StringVar(&dbHost, "db-host", "localhost", "database host")
	flag.StringVar(&dbPort, "db-port", "5432", "database port")
	flag.StringVar(&dbUser, "db-user", "", "database user")
	flag.StringVar(&dbPassword, "db-password", "", "database password")
	flag.StringVar(&dbName, "db-name", "", "database name")
	flag.StringVar(&sslMode, "ssl-mode", "disable", "SSL mode (disable, require, verify-full, etc.)")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse()
	if dbUser == "" {
		panic("db-user is required")
	}
	if dbName == "" {
		panic("db-name is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&x-migrations-table=%s",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
		sslMode,
		migrationsTable,
	)

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}

		panic(err)
	}

	fmt.Println("migrations applied successfully")
}
