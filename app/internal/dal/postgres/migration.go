package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"software_test/internal/config"
	"software_test/internal/dal/postgres/migrations"
)

func RunMigrations(cfg *config.PostgresConfig) error {
	stdlib.GetDefaultDriver()

	pgDsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	if cfg.Binary {
		pgDsn += "?sslmode=require"
	}

	db, err := goose.OpenDBWithDriver("pgx", pgDsn)
	if err != nil {
		return err
	}

	goose.SetBaseFS(migrations.EmbedMigrations)

	err = goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	err = goose.Up(db, ".")
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
	}

	return nil
}
