package utils

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

func TxToRLPString(tx *types.Transaction) (string, error) {
	b, err := tx.MarshalBinary()
	if err != nil {
		return "", err
	}
	return hexutil.Encode(b), nil
}
