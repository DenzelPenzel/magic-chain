package process

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/denzelpenzel/magic-chain/internal/client"
	"github.com/denzelpenzel/magic-chain/internal/core"
	"github.com/denzelpenzel/magic-chain/internal/logging"
	"github.com/denzelpenzel/magic-chain/internal/state"
	"github.com/denzelpenzel/magic-chain/internal/utils"
	ethcore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/rpc"
	"go.uber.org/zap"
)

type Routine interface {
	Loop(ctx context.Context, processChan chan *types.Transaction) (*rpc.ClientSubscription, error)
	Height() (*big.Int, error)
}
type ChainReader struct {
	ctx context.Context

	routine   Routine
	jobEvents chan core.Event
	close     chan int
	store     *state.FileStore

	wg *sync.WaitGroup
}

func NewReader(ctx context.Context, r Routine, store *state.FileStore) (Process, error) {
	cr := &ChainReader{
		ctx:       ctx,
		routine:   r,
		jobEvents: make(chan core.Event, 100),
		wg:        &sync.WaitGroup{},
		close:     make(chan int),
		store:     store,
	}

	return cr, nil
}

func (cr *ChainReader) Close() error {
	cr.close <- killSig
	cr.wg.Wait()
	return nil
}

func (cr *ChainReader) EventLoop() error {
	logger := logging.WithContext(cr.ctx)
	logger.Debug("Starting process job")

	jobCtx, cancel := context.WithCancel(cr.ctx)

	cr.wg.Add(1)

	go func() {
		defer cr.wg.Done()

		localTx := make(chan *types.Transaction)

		sub, err := cr.routine.Loop(cr.ctx, localTx)
		if err != nil {
			logger.Error("Received error from read routine", zap.Error(err))
			return
		}

		for {
			select {
			case err = <-sub.Err():
				logger.Error("Subscription error.", zap.Error(err))
				return

			case tx := <-localTx:
				cr.jobEvents <- core.Event{Timestamp: time.Now().UTC(), Value: tx}

			case <-jobCtx.Done():
				return
			}
		}
	}()

	for {
		select {
		case event := <-cr.jobEvents:
			logger.Info("Received the new event", zap.Any("event", event))
			cr.processTx(event)

		case <-cr.close:
			logger.Debug("Shutting down reader process")
			cancel()
			return nil
		}
	}
}

func (cr *ChainReader) processTx(event core.Event) {
	logger := logging.WithContext(cr.ctx)

	tx := event.Value
	txHashLower := strings.ToLower(tx.Hash().Hex())

	logger.Debug("Processing tx", zap.String("txHash", txHashLower))

	outFiles, err := cr.store.GetCSVFile(event.Timestamp.Unix())
	if err != nil {
		logger.Error("Failed to get CSV file", zap.Error(err))
		return
	}

	logger.Info("Store tx in CSV file",
		zap.String("source", outFiles.FSourcelog.Name()),
		zap.String("tx", outFiles.FTxs.Name()),
	)

	_, err = fmt.Fprintf(outFiles.FSourcelog, "%d,%s,%s\n", event.Timestamp.UnixMilli(), txHashLower, event.Source)
	if err != nil {
		logger.Error("Failed to store fsourcelog data", zap.Error(err))
		return
	}

	_, err = cr.store.GetTx(txHashLower)
	if err == nil {
		logger.Error("Transaction already processed")
		return
	}

	if err := cr.validateTx(event); err != nil {
		return
	}

	l1Client, err := client.FromNetwork(cr.ctx)
	if err != nil {
		return
	}

	receipt, err := l1Client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		if err.Error() != "not found" {
			logger.Error("Failed to execute ethClient.TransactionReceipt", zap.Error(err))
		}
	}

	if receipt != nil {
		logger.Info("Tx already included", zap.Uint64("block", receipt.BlockNumber.Uint64()))
		return
	}

	rlpHex, err := utils.TxToRLPString(event.Value)
	if err != nil {
		logger.Error("Failed to encode rlp", zap.Error(err))
		return
	}

	_, err = fmt.Fprintf(outFiles.FTxs, "%d,%s,%s\n", event.Timestamp.UnixMilli(), txHashLower, rlpHex)
	if err != nil {
		logger.Error("Failed to store tx in file", zap.Error(err))
		return
	}

	_, err = cr.store.SetTx(txHashLower, event.Timestamp)
	if err != nil {
		logger.Error("Failed to store tx", zap.Error(err))
		return
	}
}

func (cr *ChainReader) validateTx(event core.Event) error {
	tx := event.Value

	if _, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx); err != nil {
		return err
	}

	if tx.Value().Sign() < 0 {
		return txpool.ErrNegativeValue
	}

	if tx.GasFeeCap().BitLen() > 256 {
		return ethcore.ErrFeeCapVeryHigh
	}

	if tx.GasTipCap().BitLen() > 256 {
		return ethcore.ErrTipVeryHigh
	}

	if tx.GasFeeCapIntCmp(tx.GasTipCap()) < 0 {
		return ethcore.ErrTipAboveFeeCap
	}

	return nil
}
