package shared

import (
	"context"
	"database/sql"
	"io"

	"github.com/rs/zerolog"
	"github.com/vkumov/go-pxgrider/server/internal/connection"
)

type (
	Logger interface {
		Logger() *zerolog.Logger
		LoggerWriter() io.Writer
	}

	DBer interface {
		DB() *sql.DB
	}

	UserHandler interface {
		LoadFromDB(ctx context.Context) error
		AddConnection(ctx context.Context, req connection.ConnectionCreate) (*connection.Connection, error)
		GetConnection(id string) (*connection.Connection, error)
		FindConnection(name string) (*connection.Connection, error)
		DeleteConnection(ctx context.Context, id string) error

		GetConnections() []*connection.Connection
	}

	UsersHandler interface {
		GetUser(context.Context, string) UserHandler
	}

	App interface {
		Start() error
		Log() *zerolog.Logger
		Users() UsersHandler
		Ready() <-chan struct{}
	}
)
