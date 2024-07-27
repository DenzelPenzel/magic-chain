package client

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
)

func NewEthClient(ctx context.Context, rawURL string) (*ethclient.Client, error) {
	return ethclient.DialContext(ctx, rawURL)
}
