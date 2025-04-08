package db

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"time"
)

type (
	DBSpecs struct {
		ConnRetry             int           `default:"1"`
		ConnTimeout           time.Duration `default:"1m"`
		SSLMode               string
		User                  string
		Password              string
		Host                  string
		Port                  string
		Name                  string
		MaxParamsPerStatement int `default:"32767"`
	}
)

func (cfg *DBSpecs) Validate() error {
	if cfg.User == "" {
		return errors.New("no DB user provided")
	}
	if cfg.Host == "" {
		return errors.New("no DB host provided")
	}
	if cfg.Port == "" {
		return errors.New("no DB port provided")
	}
	if cfg.Name == "" {
		return errors.New("no DB name provided")
	}
	return nil
}

func (cfg DBSpecs) GetDSN() string { // nolint:gocritic
	query := make(url.Values)
	if cfg.SSLMode != "" {
		query.Set("sslmode", cfg.SSLMode)
	}
	if cfg.ConnTimeout > 0 {
		query.Set("connect_timeout", strconv.Itoa(int(cfg.ConnTimeout/time.Second)))
	}
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     cfg.Name,
		RawQuery: query.Encode(),
	}
	return dsn.String()
}
