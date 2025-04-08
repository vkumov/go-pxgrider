package internal

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"

	"github.com/vkumov/go-pxgrider/server/internal/connection"
	"github.com/vkumov/go-pxgrider/server/internal/db/models"
	"github.com/vkumov/go-pxgrider/server/shared"
)

type (
	user struct {
		uid         string
		connections map[string]*connection.Connection
		l           *zerolog.Logger
		lw          io.Writer
		db          *sql.DB
		lock        sync.RWMutex
	}
)

var _ shared.UserHandler = (*user)(nil)

func newUser(ctx context.Context, uid string, l *zerolog.Logger, db *sql.DB, lw io.Writer) *user {
	u := &user{
		uid:         uid,
		l:           l,
		lw:          lw,
		db:          db,
		connections: make(map[string]*connection.Connection),
	}

	//FIXME: re-think which context to use
	l.Debug().Str("uid", uid).Msg("Creating new user")
	if err := u.LoadFromDB(ctx); err != nil {
		l.Error().Err(err).Msg("Failed to load user from db")
	}

	return u
}

func (u *user) LoadFromDB(ctx context.Context) error {
	u.l.Debug().Str("uid", u.uid).Msg("Loading user from db")

	cls, err := models.Clients(models.ClientWhere.Owner.EQ(u.uid)).All(ctx, u.db)
	if err != nil {
		return fmt.Errorf("failed to load clients for user %s: %w", u.uid, err)
	}

	u.lock.Lock()
	defer u.lock.Unlock()

	for _, cl := range cls {
		c := connection.New(u.db, cl.ID, u.uid, u.l, u.lw)
		if err := c.WithDBData(cl); err != nil {
			return fmt.Errorf("failed to load connection %s for user %s: %w", cl.ID, u.uid, err)
		}
		u.connections[cl.ID] = c
	}

	return nil
}

func (u *user) AddConnection(ctx context.Context, req connection.ConnectionCreate) (*connection.Connection, error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate connection id: %w", err)
	}

	c, err := connection.NewWithRequest(u.db, id.String(), u.uid, req, u.l, u.lw)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	if err := c.RebuildPxGridConfig(); err != nil {
		return nil, fmt.Errorf("failed to rebuild pxgrid config: %w", err)
	}

	if err = c.Store(ctx); err != nil {
		return c, fmt.Errorf("failed to store connection: %w", err)
	}
	u.connections[c.ID()] = c

	return c, nil
}

func (u *user) GetConnection(id string) (*connection.Connection, error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	if id == "" {
		return nil, fmt.Errorf("connection id is empty")
	}

	c, ok := u.connections[id]
	if !ok {
		return nil, fmt.Errorf("connection %s not found", id)
	}

	return c, nil
}

func (u *user) FindConnection(name string) (*connection.Connection, error) {
	u.lock.Lock()
	defer u.lock.Unlock()

	if name == "" {
		return nil, fmt.Errorf("connection name is empty")
	}

	for _, c := range u.connections {
		if c.Name() == name {
			return c, nil
		}
	}

	return nil, fmt.Errorf("connection %s not found", name)
}

func (u *user) GetConnections() []*connection.Connection {
	u.lock.RLock()
	defer u.lock.RUnlock()

	res := make([]*connection.Connection, 0, len(u.connections))
	for _, c := range u.connections {
		res = append(res, c)
	}

	return res
}

func (u *user) DeleteConnection(ctx context.Context, id string) error {
	fail := func(err error) error {
		return fmt.Errorf("failed to delete connection %s: %w", id, err)
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	if _, err := models.Logs(models.LogWhere.Client.EQ(id)).DeleteAll(ctx, tx); err != nil {
		return fail(err)
	}

	if _, err := models.Messages(models.MessageWhere.Client.EQ(id)).DeleteAll(ctx, tx); err != nil {
		return fail(err)
	}

	if _, err := models.Clients(models.ClientWhere.ID.EQ(id)).DeleteAll(ctx, tx); err != nil {
		return fail(err)
	}

	if err := tx.Commit(); err != nil {
		return fail(err)
	}

	u.lock.Lock()
	defer u.lock.Unlock()

	c, ok := u.connections[id]
	if !ok {
		return nil
	}

	c.Stop()

	delete(u.connections, id)

	return nil
}
