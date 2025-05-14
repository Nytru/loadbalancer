package datebase

import (
	"cloud-test/internal/configuration"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func MigrateUp(cfg configuration.PgDbCfg) error {
	m, err := migrate.New(
		getMigrationPath(cfg),
		getMigrationString(cfg),
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func GetPostgresConnectionString(cfg configuration.PgDbCfg) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName)
}

func getMigrationString(cfg configuration.PgDbCfg) string {
	return fmt.Sprintf("pgx5://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
	)
}

func getMigrationPath(cfg configuration.PgDbCfg) string {
	return fmt.Sprintf("file://%s", cfg.MigrationsPath)
}
