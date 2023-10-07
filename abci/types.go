package abci

import (
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/mempool"
)

type PrepareProposalHandler struct {
	logger      log.Logger
	txConfig    client.TxConfig
	cdc         codec.Codec
	mempool     *mempool.SenderNonceMempool
}

type ProcessProposalHandler struct {
	TxConfig client.TxConfig
	Codec    codec.Codec
	Logger   log.Logger
}

func NewPrepareProposalHandler(
	logger log.Logger, txConfig client.TxConfig, cdc codec.Codec, mempool *mempool.SenderNonceMempool,
) *PrepareProposalHandler {
	return &PrepareProposalHandler{
		logger:      logger,
		txConfig:    txConfig,
		cdc:         cdc,
		mempool:     mempool,
	}
}
