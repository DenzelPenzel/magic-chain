package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/denzelpenzel/magic-chain/internal/config"
	"github.com/denzelpenzel/magic-chain/internal/logging"
	"github.com/denzelpenzel/magic-chain/internal/manager"
	"go.uber.org/zap"
)

type Application struct {
	cfg *config.Config
	ctx context.Context
	m   *manager.Manager
}

func New(ctx context.Context, cfg *config.Config, m *manager.Manager) *Application {
	return &Application{
		ctx: ctx,
		cfg: cfg,
		m:   m,
	}
}

func (a *Application) Start() error {
	a.m.StartEventRoutines(a.ctx)

	if err := a.m.Run(); err != nil {
		return err
	}

	return nil
}

// ListenForShutdown handles and listens for shutdown
func (a *Application) ListenForShutdown(stop func()) {
	done := <-a.End() // Blocks until an OS signal is received
	logging.WithContext(a.ctx).Info("Received shutdown OS signal", zap.String("signal", done.String()))
	stop()
}

// End returns a channel that will receive an OS signal
func (a *Application) End() <-chan os.Signal {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	return sigs
}
