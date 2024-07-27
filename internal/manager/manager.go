package manager

import (
	"context"
	"sync"

	"github.com/denzelpenzel/magic-chain/internal/config"
	"github.com/denzelpenzel/magic-chain/internal/etl"
	"github.com/denzelpenzel/magic-chain/internal/logging"
	"github.com/denzelpenzel/magic-chain/internal/process"
	"go.uber.org/zap"
)

type Manager struct {
	ctx     context.Context
	cfg     *config.Config
	etl     etl.ETL
	process process.Process

	*sync.WaitGroup
}

func NewManager(ctx context.Context, cfg *config.Config, etl etl.ETL) *Manager {
	return &Manager{
		ctx:       ctx,
		cfg:       cfg,
		etl:       etl,
		WaitGroup: &sync.WaitGroup{},
	}
}

func (m *Manager) StartEventRoutines(ctx context.Context) {
	logger := logging.WithContext(ctx)

	m.Add(1)

	go func() {
		defer m.Done()
		if err := m.etl.EventLoop(); err != nil {
			logger.Error("engine manager event loop error", zap.Error(err))
		}
	}()
}

// EventLoop ... Driver ran as separate go routine
func (m *Manager) EventLoop() error {
	logger := logging.WithContext(m.ctx)
	for {
		<-m.ctx.Done()
		logger.Info("Shutting down ETL")
		return nil
	}
}

func (m *Manager) Shutdown() error {
	if err := m.etl.Shutdown(m.process); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Run() error {
	p, err := m.etl.CreateProcess(m.cfg)
	if err != nil {
		return err
	}
	m.process = p
	return m.etl.Run(p)
}
