package auth

import (
	"context"
	"crypto/subtle"
	"errors"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	ErrEmptyMetadata = errors.New("metadata is empty")
	ErrAccessDenied  = errors.New("access denied")
)

type (
	tokenAuth struct {
		token string
		log   *zerolog.Logger
	}

	Authenticator interface {
		StreamInterceptor(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error
		UnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error)
	}
)

var _ Authenticator = (*tokenAuth)(nil)

func NewTokenAuthenticator(token string, log *zerolog.Logger) Authenticator {
	a := &tokenAuth{log: log, token: token}
	return a
}

func (a *tokenAuth) StreamInterceptor(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := a.authorize(stream.Context()); err != nil {
		return err
	}

	a.log.Debug().Msg("stream interceptor")
	return handler(srv, stream)
}

func (a *tokenAuth) UnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if err := a.authorize(ctx); err != nil {
		return nil, err
	}

	a.log.Debug().Msg("unary interceptor")
	return handler(ctx, req)
}

func (a tokenAuth) authorize(ctx context.Context) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		token, ok := md["token"]
		if !ok || len(token) == 0 {
			return ErrAccessDenied
		}

		if subtle.ConstantTimeCompare([]byte(token[0]), []byte(a.token)) == 1 {
			return nil
		}

		return ErrAccessDenied
	}

	return ErrEmptyMetadata
}
