package logger

import (
	"context"

	"github.com/rs/zerolog"
	gopxgrid "github.com/vkumov/go-pxgrid"
)

type likeZerolog interface {
	Debug() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	Info() *zerolog.Event
	Log() *zerolog.Event
	Panic() *zerolog.Event
	Trace() *zerolog.Event
	Warn() *zerolog.Event
	With() zerolog.Context
}

type PxGridLog struct {
	Logger likeZerolog
}

var _ gopxgrid.Logger = (*PxGridLog)(nil)

func (l *PxGridLog) Debug(msg string, args ...any) {
	l.Logger.Debug().Fields(args).Msg(msg)
}

func (l *PxGridLog) DebugContext(ctx context.Context, msg string, args ...any) {
	l.Logger.Debug().Ctx(ctx).Fields(args).Msg(msg)
}

func (l *PxGridLog) Error(msg string, args ...any) {
	l.Logger.Error().Fields(args).Msg(msg)
}

func (l *PxGridLog) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.Logger.Error().Ctx(ctx).Fields(args).Msg(msg)
}

func (l *PxGridLog) Info(msg string, args ...any) {
	l.Logger.Info().Fields(args).Msg(msg)
}

func (l *PxGridLog) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Logger.Info().Ctx(ctx).Fields(args).Msg(msg)
}

func (l *PxGridLog) Warn(msg string, args ...any) {
	l.Logger.Warn().Fields(args).Msg(msg)
}

func (l *PxGridLog) WarnContext(ctx context.Context, msg string, args ...any) {
	l.Logger.Warn().Ctx(ctx).Fields(args).Msg(msg)
}

func (l *PxGridLog) With(args ...any) gopxgrid.Logger {
	n := l.Logger.With().Fields(args).Logger()
	return &PxGridLog{&n}
}
