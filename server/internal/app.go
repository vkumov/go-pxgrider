package internal

import (
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"

	pb "github.com/vkumov/go-pxgrider/pkg"

	"github.com/vkumov/go-pxgrider/server/internal/auth"
	"github.com/vkumov/go-pxgrider/server/internal/config"
	"github.com/vkumov/go-pxgrider/server/internal/server"
	"github.com/vkumov/go-pxgrider/server/shared"
)

type App struct {
	cfg        *config.AppConfig
	users      shared.UsersHandler
	grpcServer *grpc.Server
	pxServer   pb.PxgriderServiceServer
	health     *health.Server

	ready chan struct{}
}

func NewApp() *App {
	app := &App{
		cfg:   config.NewConfig(),
		ready: make(chan struct{}),
	}
	app.users = NewUsers(app.cfg, app.cfg)

	authLogger := app.cfg.Logger().With().Str("component", "auth").Logger()
	aut := auth.NewTokenAuthenticator(app.cfg.Token(), &authLogger)

	app.grpcServer = grpc.NewServer(
		grpc.StreamInterceptor(aut.StreamInterceptor),
		grpc.UnaryInterceptor(aut.UnaryInterceptor),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             app.cfg.Specs.Server.EnforcementPolicy.MinTime,
			PermitWithoutStream: app.cfg.Specs.Server.EnforcementPolicy.PermitWithoutStream,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     app.cfg.Specs.Server.Keepalive.MaxConnectionIdle,
			MaxConnectionAge:      app.cfg.Specs.Server.Keepalive.MaxConnectionAge,
			MaxConnectionAgeGrace: app.cfg.Specs.Server.Keepalive.MaxConnectionAgeGrace,
			Time:                  app.cfg.Specs.Server.Keepalive.Time,
			Timeout:               app.cfg.Specs.Server.Keepalive.Timeout,
		}),
	)
	app.pxServer = server.NewServer(app)
	pb.RegisterPxgriderServiceServer(app.grpcServer, app.pxServer)

	app.health = health.NewServer()
	healthgrpc.RegisterHealthServer(app.grpcServer, app.health)

	return app
}

func (a *App) Start() error {
	listen := fmt.Sprintf(":%d", a.cfg.Specs.Server.Port)
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", a.cfg.Specs.Server.Port, err)
	}

	a.cfg.Logger().Info().
		Str("build_stamp", a.cfg.Specs.Version.BuildStamp).
		Str("git_hash", a.cfg.Specs.Version.GitHash).
		Str("git_version", a.cfg.Specs.Version.GitVersion).
		Str("version", a.cfg.Specs.Version.V).
		Str("env", a.cfg.Specs.Env).
		Str("address", listen).
		Msg("Starting server")

	close(a.ready)

	a.health.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	return a.grpcServer.Serve(lis)
}

func (a *App) Log() *zerolog.Logger {
	return a.cfg.Logger()
}

func (a *App) Users() shared.UsersHandler {
	return a.users
}

func (a *App) Ready() <-chan struct{} {
	return a.ready
}

func (a *App) IsProd() bool {
	return a.cfg.IsProd()
}
