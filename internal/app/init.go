package app

import (
	"context"

	"github.com/denzelpenzel/magic-chain/internal/client"
	"github.com/denzelpenzel/magic-chain/internal/config"
	"github.com/denzelpenzel/magic-chain/internal/core"
	"github.com/denzelpenzel/magic-chain/internal/etl"
	"github.com/denzelpenzel/magic-chain/internal/logging"
	"github.com/denzelpenzel/magic-chain/internal/manager"
	"github.com/denzelpenzel/magic-chain/internal/registry"
	"go.uber.org/zap"
)

func InitContext(ctx context.Context, cb *client.Bundle) context.Context {
	return context.WithValue(ctx, core.Clients, cb)
}

func NewMagicChainApp(ctx context.Context, cfg *config.Config) (*Application, func(), error) {
	r := registry.New()
	e := etl.New(ctx, r)
	m := manager.NewManager(ctx, cfg, e)

	appShutDown := func() {
		if err := m.Shutdown(); err != nil {
			logging.WithContext(ctx).Error("error shutting down subsystems", zap.Error(err))
		}
	}

	return New(ctx, cfg, m), appShutDown, nil
}
