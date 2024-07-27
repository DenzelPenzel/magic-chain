package client

import (
	"context"
	"fmt"

	"github.com/denzelpenzel/magic-chain/internal/core"
	"github.com/denzelpenzel/magic-chain/internal/logging"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"
)

type Bundle struct {
	L1Client     *ethclient.Client
	L1NodeClient *gethclient.Client
}

type Config struct {
	L1RpcEndpoint string
}

func NewNodeClient(rawURL string) (*gethclient.Client, error) {
	rpcClient, err := rpc.Dial(rawURL)
	if err != nil {
		return nil, err
	}
	logging.NoContext().Debug("Successfully connected to node", zap.String("URL", rawURL))
	return gethclient.New(rpcClient), nil
}

func NewBundle(ctx context.Context, cfg *core.ClientConfig) (*Bundle, error) {
	logger := logging.WithContext(ctx)

	l1Client, err := NewEthClient(ctx, cfg.L1RpcEndpoint)
	if err != nil {
		logger.Fatal("Error creating L1 client", zap.Error(err))
		return nil, err
	}

	l1NodeClient, err := NewNodeClient(cfg.L1RpcEndpoint)
	if err != nil {
		logger.Fatal("Error creating L1 node client", zap.Error(err))
		return nil, err
	}

	return &Bundle{
		L1Client:     l1Client,
		L1NodeClient: l1NodeClient,
	}, nil
}

func FromContext(ctx context.Context) (*Bundle, error) {
	b, err := ctx.Value(core.Clients).(*Bundle)
	if !err {
		return nil, fmt.Errorf("failed to retrieve client bundle from context")
	}
	return b, nil
}

func (b *Bundle) NodeClient() *gethclient.Client {
	return b.L1NodeClient
}

func FromNetwork(ctx context.Context) (*ethclient.Client, error) {
	bundle, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return bundle.L1Client, nil
}
