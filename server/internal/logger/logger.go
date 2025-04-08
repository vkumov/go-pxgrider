package logger

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vkumov/go-pxgrider/server/internal/db/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type (
	combinedWriter struct {
		originalWriter io.Writer
		db             *sql.DB
		eventStream    chan map[string]interface{}
	}

	Logger struct {
		*zerolog.Logger
		w *combinedWriter
	}
)

const (
	ConnectionIdFieldName = "connection_id"
	ComponentFieldName    = "component"
	UsernameFieldName     = "username"

	queueSize = 32
)

func NewCombined(connectionId string, fromLogger *zerolog.Logger, db *sql.DB, writer io.Writer, fields ...any) *Logger {
	builder := fromLogger.With().Str(ConnectionIdFieldName, connectionId)

	if len(fields) > 0 {
		builder = builder.Fields(fields)
	}

	stdoutLogger := builder.Logger()
	wr := newCombinedWriter(db, writer)
	dbLogger := stdoutLogger.Output(wr)

	return &Logger{
		Logger: &dbLogger,
		w:      wr,
	}
}

func newCombinedWriter(db *sql.DB, writer io.Writer) *combinedWriter {
	c := &combinedWriter{
		db:             db,
		originalWriter: writer,
		eventStream:    make(chan map[string]any, queueSize),
	}

	go c.storeEvent()

	return c
}

func (w *combinedWriter) Write(p []byte) (n int, err error) {
	n, err = w.originalWriter.Write(p)
	if err != nil {
		return
	}

	var evt map[string]interface{}
	p = decodeIfBinaryToBytes(p)
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err = d.Decode(&evt)
	if err != nil {
		return
	}
	w.eventStream <- evt

	return
}

func (w *combinedWriter) Stop() {
	close(w.eventStream)
}

func (l *Logger) Stop() {
	l.w.Stop()
}

func (l *Logger) Level(level string) error {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return err
	}
	newLogger := l.Logger.Level(lvl)
	l.Logger = &newLogger
	return nil
}

func (w *combinedWriter) storeEvent() {
	for evt := range w.eventStream {
		connectionId, ok := evt[ConnectionIdFieldName].(string)
		if !ok {
			log.Error().Err(fmt.Errorf("cannot extract connection_id from event")).Send()
			continue
		}

		timestamp, ok := evt[zerolog.TimestampFieldName].(string)
		if !ok {
			log.Error().Err(fmt.Errorf("cannot extract timestamp from event")).Interface("raw", evt).Send()
			continue
		}

		level, ok := evt[zerolog.LevelFieldName].(string)
		if !ok {
			log.Error().Err(fmt.Errorf("cannot extract level from event")).Send()
			continue
		}

		var label null.String
		component, ok := evt[ComponentFieldName].(string)
		if ok {
			label = null.StringFrom(component)
		}

		t, err := time.Parse(time.RFC3339Nano, timestamp)
		if err != nil {
			log.Error().Err(fmt.Errorf("cannot parse timestamp: %s", err)).Send()
			continue
		}

		cleanupEvent(evt)
		message, err := json.Marshal(evt)
		if err != nil {
			log.Error().Err(fmt.Errorf("cannot marshal event: %s", err)).Send()
			continue
		}

		l := models.Log{
			Client:    connectionId,
			Level:     level,
			Timestamp: null.TimeFrom(t),
			Message:   null.StringFrom(string(message)),
			Label:     label,
		}
		err = l.Insert(context.Background(), w.db, boil.Infer())
		if err != nil {
			log.Error().Err(err).Msg("failed to insert log into db")
		}
	}
}

func cleanupEvent(evt map[string]interface{}) {
	delete(evt, ConnectionIdFieldName)
	delete(evt, ComponentFieldName)
	delete(evt, zerolog.TimestampFieldName)
	delete(evt, zerolog.LevelFieldName)
	delete(evt, zerolog.CallerFieldName)
	delete(evt, UsernameFieldName)
}

func decodeIfBinaryToBytes(in []byte) []byte {
	return in
}
