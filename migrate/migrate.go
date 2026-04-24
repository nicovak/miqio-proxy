package migrate

import (
	"context"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5"
	cfg "github.com/nicovak/miqio-proxy/config"

	"github.com/nicovak/miqio-go/config"
	"github.com/nicovak/miqio-go/driver"
	"github.com/nicovak/miqio-go/logger"
)

func removeParam(connString, param string) (string, error) {
	u, err := url.Parse(connString)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Del(param)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func getVersion() int {
	query := `
		SELECT version
		FROM schema_migrations
	`

	version := 0

	postgres := driver.NewPostgres(cfg.Get().Postgres.ConnectionString)
	row := postgres.ConnPool.QueryRow(context.Background(), query)

	err := row.Scan(&version)

	if err != nil && err != pgx.ErrNoRows {
		logger.Get().Panicf("migrate - error from DB : %s", err)
	}

	return version
}

func Run() {
	connString := cfg.Get().Postgres.ConnectionString

	connString, err := removeParam(connString, "pool_max_conns")
	if err != nil {
		logger.Get().Panicln("migrate", err)
	}

	m, err := migrate.New(
		`file://`+config.RootPath+`db/migrations`,
		connString,
	)
	if err != nil {
		logger.Get().Panicln("migrate", err)
	}

	version := getVersion()

	if errUp := m.Up(); errUp != migrate.ErrNoChange {
		if version != 0 {
			if errForce := m.Force(version); errForce != nil {
				logger.Get().Panicln("migrate", err)
			}

			return
		}

		logger.Get().Panicln("migrate", errUp)
	}
}
