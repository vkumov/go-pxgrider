package internal

import (
	"context"
	"database/sql"
	"io"
	"sync"

	"github.com/rs/zerolog"

	"github.com/vkumov/go-pxgrider/server/internal/logger"
	"github.com/vkumov/go-pxgrider/server/shared"
)

type Users struct {
	users map[string]shared.UserHandler
	l     *zerolog.Logger
	lw    io.Writer
	db    *sql.DB
	lock  sync.Mutex
}

var _ shared.UsersHandler = (*Users)(nil)

func (u *Users) GetUser(ctx context.Context, username string) shared.UserHandler {
	u.lock.Lock()
	defer u.lock.Unlock()

	if _, ok := u.users[username]; !ok {
		log := u.l.With().Str(logger.UsernameFieldName, username).Logger()
		u.users[username] = newUser(ctx, username, &log, u.db, u.lw)
	}

	return u.users[username]
}

func NewUsers(l shared.Logger, db shared.DBer) *Users {
	return &Users{
		users: make(map[string]shared.UserHandler),
		l:     l.Logger(),
		lw:    l.LoggerWriter(),
		db:    db.DB(),
	}
}
