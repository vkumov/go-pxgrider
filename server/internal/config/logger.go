package config

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/vkumov/go-pxgrider/server/shared"
)

var _ (shared.Logger) = (*AppConfig)(nil)

func (app *AppConfig) buildLogger() *AppConfig {
	// preparing logger
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.ErrorFieldName = "err"

	level := configLevelToZerologLevel(app.Specs.Log.Level)
	var l zerolog.Logger
	if app.IsProd() {
		// pure JSON
		app.logWriter = os.Stderr
		l = zerolog.New(app.logWriter).
			With().
			Timestamp().
			Logger().
			Level(level)

		zerolog.SetGlobalLevel(level)
	} else {
		// with colored output
		app.logWriter = zerolog.ConsoleWriter{Out: os.Stderr}
		l = zerolog.New(app.logWriter).
			With().
			Caller().
			Timestamp().
			Logger().
			Level(level)
		l.Debug().Msg("Starting app with DEBUGs enabled")

		log.Logger = zerolog.New(app.logWriter).
			With().
			Timestamp().
			Logger().
			Level(level)
	}
	app.l = &l
	return app
}

func (app *AppConfig) Logger() *zerolog.Logger {
	return app.l
}

func (app *AppConfig) LoggerWriter() io.Writer {
	return app.logWriter
}

func configLevelToZerologLevel(level string) zerolog.Level {
	v, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.InfoLevel
	}
	return v
}
