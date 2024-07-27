package registry

import (
	"context"
	"math/big"

	"github.com/denzelpenzel/magic-chain/internal/client"
	"github.com/denzelpenzel/magic-chain/internal/config"
	"github.com/denzelpenzel/magic-chain/internal/process"
	"github.com/denzelpenzel/magic-chain/internal/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type NodeTraversal struct {
	nodeClient *gethclient.Client
}

func NewHeaderTraversal(ctx context.Context, cfg *config.Config) (process.Process, error) {
	clients, err := client.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	nodeClient := clients.NodeClient()
	store := state.NewFileStore(cfg.DataDir)

	// remove old transactions in background
	go store.Cleaner()

	nt := &NodeTraversal{
		nodeClient: nodeClient,
	}

	reader, err := process.NewReader(ctx, nt, store)
	if err != nil {
		return nil, err
	}

	return reader, err
}

func (ht *NodeTraversal) Loop(ctx context.Context, consumer chan *types.Transaction) (*rpc.ClientSubscription, error) {
	sub, err := ht.nodeClient.SubscribeFullPendingTransactions(ctx, consumer)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func (ht *NodeTraversal) Height() (*big.Int, error) {
	// TODO implement me
	panic("implement me")
}
