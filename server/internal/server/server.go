package server

import (
	"context"
	"errors"

	pb "github.com/vkumov/go-pxgrider/pkg"

	"github.com/vkumov/go-pxgrider/server/internal/connection"
	"github.com/vkumov/go-pxgrider/server/shared"
)

type server struct {
	app shared.App
	pb.UnimplementedPxgriderServiceServer
}

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

var _ pb.PxgriderServiceServer = (*server)(nil)

func NewServer(app shared.App) pb.PxgriderServiceServer {
	return &server{app: app}
}

func (s *server) getUserConnection(ctx context.Context, uid string, connectionID string) (shared.UserHandler, *connection.Connection, error) {
	u := s.app.Users().GetUser(ctx, uid)
	if u == nil {
		return nil, nil, ErrUserNotFound
	}

	c, err := u.GetConnection(connectionID)
	if err != nil {
		return nil, nil, err
	}

	return u, c, nil
}
