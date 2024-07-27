package process

import (
	"context"

	"github.com/denzelpenzel/magic-chain/internal/config"
)

const (
	killSig = 0
)

type Process interface {
	Close() error
	EventLoop() error
}

type (
	Constructor = func(context.Context, *config.Config) (Process, error)
)
