package config

import (
	"database/sql"
	"flag"
	"io"
	"sync"

	"github.com/rs/zerolog"
)

type AppConfig struct {
	sync.RWMutex

	Specs Specs

	db        *sql.DB
	logWriter io.Writer
	l         *zerolog.Logger
}

func NewConfig() *AppConfig {
	loadEnv()

	cfgFile := flag.String("config", "", "specifies config file to use")
	flag.Parse()

	a := &AppConfig{}
	a.mustLoadSpecs(cfgFile).
		buildLogger().
		mustInitDB()

	return a
}

func (c *AppConfig) Token() string {
	c.RLock()
	defer c.RUnlock()

	return c.Specs.Auth.Token
}

func (c *AppConfig) IsProd() bool {
	c.RLock()
	defer c.RUnlock()

	return c.Specs.Env == "prod"
}
