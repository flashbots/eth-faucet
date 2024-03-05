package txbuilder

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/flashbots/eth-faucet/backoff"
	"github.com/flashbots/eth-faucet/config"
	"github.com/flashbots/eth-faucet/logutils"
	"go.uber.org/zap"
)

var (
	ErrFailedToConnectToRPC    = errors.New("failed to connect to rpc")
	ErrFailedToReadChainID     = errors.New("failed to read chain id from rpc")
	ErrFailedToRefreshNonce    = errors.New("failed to refresh nonce")
	ErrFailedToSendTransaction = errors.New("failed to send transaction")
	ErrFailedToSignTransaction = errors.New("failed to sign transaction")
	ErrFailedToSuggestGasPrice = errors.New("failed to suggest gas price")
)

type TxBuilder struct {
	address       common.Address
	backoffParams *backoff.Parameters
	client        bind.ContractTransactor
	privateKey    *ecdsa.PrivateKey
	signer        types.Signer

	nonce uint64
}

func New(cfg *config.Config) (*TxBuilder, error) {
	l := zap.L()

	privateKey, err := cfg.Wallet.ECDSA()
	if err != nil {
		return nil, err
	}

	backoffParams := &backoff.Parameters{
		BaseTimeout: cfg.RPC.Timeout,
	}

	l.Info("Connecting to rpc endpoint...", zap.String("rpc_endpoint", cfg.RPC.Endpoint))
	var client *ethclient.Client
	err = backoff.Backoff(context.Background(), backoffParams, func(_ context.Context) (_err error) {
		client, _err = ethclient.Dial(cfg.RPC.Endpoint)
		if _err != nil {
			l.Warn("Failed to connect to rpc endpoint", zap.Error(_err))
		}
		return _err
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToConnectToRPC, err)
	}

	var chainID *big.Int
	err = backoff.Backoff(context.Background(), backoffParams, func(ctx context.Context) (_err error) {
		chainID, _err = client.ChainID(ctx)
		if _err != nil {
			l.Warn("Failed to get chain id", zap.Error(_err))
		}
		return
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToReadChainID, err)
	}

	tb := &TxBuilder{
		address:       crypto.PubkeyToAddress(privateKey.PublicKey),
		backoffParams: backoffParams,
		client:        client,
		privateKey:    privateKey,
		signer:        types.NewEIP155Signer(chainID),
	}
	tb.refreshNonce(logutils.ContextWithLogger(context.Background(), zap.L()))

	return tb, nil
}

func (tb *TxBuilder) Address() string {
	return tb.address.String()
}

func (tb *TxBuilder) SendFunds(ctx context.Context, to string, amount *big.Int) (common.Hash, error) {
	l := logutils.LoggerFromContext(ctx)

	gasLimit := uint64(21000)

	var gasPrice *big.Int
	err := backoff.Backoff(ctx, tb.backoffParams, func(ctx context.Context) (_err error) {
		gasPrice, _err = tb.client.SuggestGasPrice(ctx)
		if _err != nil {
			l.Warn("Failed to get suggested gas price", zap.Error(_err))
		}
		return
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("%w: %w", ErrFailedToSuggestGasPrice, err)
	}

	address := common.HexToAddress(to)
	unsignedTx := types.NewTx(&types.LegacyTx{
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Nonce:    tb.getAndIncrementNonce(),
		To:       &address,
		Value:    amount,
	})

	signedTx, err := types.SignTx(unsignedTx, tb.signer, tb.privateKey)
	if err != nil {
		return signedTx.Hash(), fmt.Errorf("%w: %w", ErrFailedToSignTransaction, err)
	}

	err = backoff.Backoff(ctx, tb.backoffParams, func(_ctx context.Context) (_err error) {
		_err = tb.client.SendTransaction(_ctx, signedTx)
		if _err != nil {
			l.Warn("Failed to send transaction", zap.Error(_err), zap.String("tx_hash", signedTx.Hash().Hex()))
			if strings.Contains(_err.Error(), "nonce") {
				if _errRefresh := tb.refreshNonce(ctx); _errRefresh != nil {
					return fmt.Errorf("%w: %w", _err, _errRefresh)
				}
				_err = backoff.Retryable(_err)
			}
		}
		return
	})
	if err != nil {
		return signedTx.Hash(), fmt.Errorf("%w: %w", ErrFailedToSendTransaction, err)
	}

	return signedTx.Hash(), nil
}

func (tb *TxBuilder) refreshNonce(ctx context.Context) error {
	l := logutils.LoggerFromContext(ctx)

	var nonce uint64
	err := backoff.Backoff(context.Background(), tb.backoffParams, func(ctx context.Context) (_err error) {
		nonce, _err = tb.client.PendingNonceAt(ctx, tb.address)
		if _err != nil {
			l.Warn("Failed to refresh nonce", zap.Error(_err))
		}
		return
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToRefreshNonce, err)
	}

	tb.nonce = nonce
	return nil
}

func (tb *TxBuilder) getAndIncrementNonce() uint64 {
	return atomic.AddUint64(&tb.nonce, 1) - 1
}
