package config

import (
	"context"
	"database/sql"

	"github.com/vkumov/go-pxgrider/server/internal/db"
	"github.com/vkumov/go-pxgrider/server/shared"
)

var _ shared.DBer = (*AppConfig)(nil)

func (app *AppConfig) mustInitDB() *AppConfig {
	dblogger := app.l.With().Str("component", "db").Logger()
	db, err := db.NewSQL(context.Background(), app.Specs.DB, &dblogger)
	if err != nil {
		panic(err)
	}

	app.db = db
	return app
}

func (app *AppConfig) DB() *sql.DB {
	return app.db
}

func (app *AppConfig) DBSpec() db.DBSpecs {
	s := app.Specs.DB
	return s
}
