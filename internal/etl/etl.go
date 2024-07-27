package etl

import (
	"context"
	"fmt"
	"sync"

	"github.com/denzelpenzel/magic-chain/internal/config"
	"github.com/denzelpenzel/magic-chain/internal/core"
	"github.com/denzelpenzel/magic-chain/internal/logging"
	"github.com/denzelpenzel/magic-chain/internal/process"
	"github.com/denzelpenzel/magic-chain/internal/registry"
	"go.uber.org/zap"
)

type ETL interface {
	CreateProcess(cfg *config.Config) (process.Process, error)
	Run(p process.Process) error

	EventLoop() error
	Shutdown(p process.Process) error
}

type etl struct {
	ctx    context.Context
	cancel context.CancelFunc

	registry *registry.Registry
	wg       sync.WaitGroup
}

func New(ctx context.Context, r *registry.Registry) ETL {
	ctx, cancel := context.WithCancel(ctx)
	return &etl{
		ctx:      ctx,
		cancel:   cancel,
		registry: r,
		wg:       sync.WaitGroup{},
	}
}

func (e *etl) CreateProcess(cfg *config.Config) (process.Process, error) {
	logger := logging.WithContext(e.ctx)

	// hardcode
	dt, err := e.registry.GetDataTopic(core.BlockHeader)
	if err != nil {
		return nil, err
	}

	logger.Debug("constructing process",
		zap.String("type", dt.ProcessType.String()),
		zap.String("register_type", dt.DataType.String()))

	switch dt.ProcessType {
	case core.Subscribe:
		init, success := dt.Constructor.(process.Constructor)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(core.CouldNotCastErr, core.Read.String()))
		}

		return init(e.ctx, cfg)

	case core.Read:
		// TODO
		return nil, err

	default:
		return nil, fmt.Errorf(core.UnknownCompType, dt.ProcessType.String())
	}
}

func (e *etl) EventLoop() error {
	logger := logging.WithContext(e.ctx)

	for {
		<-e.ctx.Done()
		logger.Info("Shutting down ETL")
		return nil
	}
}

func (e *etl) Shutdown(p process.Process) error {
	e.cancel()
	logger := logging.WithContext(e.ctx)

	if err := p.Close(); err != nil {
		logger.Error("Failed to close process", zap.Error(err))
		return err
	}

	logger.Debug("Waiting for all process routines to end")
	e.wg.Wait()

	return nil
}

func (e *etl) Run(p process.Process) error {
	e.wg.Add(1)

	go func(p process.Process, wg *sync.WaitGroup) {
		defer wg.Done()
		logging.NoContext().Debug("Starting process")

		if err := p.EventLoop(); err != nil {
			logging.NoContext().Error("Obtained error from event loop", zap.Error(err))
		}
	}(p, &e.wg)

	return nil
}
