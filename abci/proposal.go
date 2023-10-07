package abci

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/srdtrk/linkedpackets"

	icacontroller "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

func (h *PrepareProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		h.logger.Info(fmt.Sprintf("🛠️ :: Prepare Proposal"))
		var proposalTxs [][]byte

		linkedPacketTxs := make(map[string]map[uint64]sdk.Tx)
		linkedPacketTotalLen := make(map[string]uint64)

		var txs []sdk.Tx
		itr := h.mempool.Select(context.Background(), nil)
		for itr != nil {
			tmptx := itr.Tx()
			sdkMsgs := tmptx.GetMsgs()

			var isLinkedTx bool
			for _, sdkMsg := range sdkMsgs {
				isLinkedPacket, linkID, linkIndex, isLastPacket := handleSdkMsg(sdkMsg)
				if isLinkedPacket {
					linkedPacketTxs[linkID][linkIndex] = tmptx
					h.logger.Info(fmt.Sprintf("🛠️ :: LinkID: %v, LinkIndex: %v", linkID, linkIndex))
					isLinkedTx = true
				}
				if isLastPacket {
					linkedPacketTotalLen[linkID] = linkIndex + 1
				}
			}

			if !isLinkedTx {
				txs = append(txs, tmptx)
			}
			itr = itr.Next()
		}
		h.logger.Info(fmt.Sprintf("🛠️ :: Number of Transactions available from mempool: %v", len(txs)))

		for _, sdkTxs := range txs {
			txBytes, err := h.txConfig.TxEncoder()(sdkTxs)
			if err != nil {
				h.logger.Info(fmt.Sprintf("❌~Error encoding transaction: %v", err.Error()))
			}
			proposalTxs = append(proposalTxs, txBytes)
		}

		for linkID, linkTxs := range linkedPacketTxs {
			if len(linkTxs) == int(linkedPacketTotalLen[linkID]) {
				for i := uint64(0); i < linkedPacketTotalLen[linkID]; i++ {
					txBytes, err := h.txConfig.TxEncoder()(linkTxs[i])
					if err != nil {
						h.logger.Info(fmt.Sprintf("❌~Error encoding transaction: %v", err.Error()))
					}
					proposalTxs = append(proposalTxs, txBytes)
				}
			} else {
				h.logger.Info(fmt.Sprintf("🛠️ :: LinkID: %v, LinkIndex: %v, TotalLen: %v", linkID, len(linkTxs), linkedPacketTotalLen[linkID]))
			}
		}

		h.logger.Info(fmt.Sprintf("🛠️ :: Number of Transactions in proposal: %v", len(proposalTxs)))

		return &abci.ResponsePrepareProposal{Txs: proposalTxs}, nil
	}
}

func handleSdkMsg(msg sdk.Msg) (bool, string, uint64, bool) {
	switch msg := msg.(type) {
	case *channeltypes.MsgRecvPacket:
		packetDataBytes := msg.Packet.GetData()

		var linkData *linkedpackets.LinkData

		var memo string
		transferPacket, err := transfer.IBCModule{}.UnmarshalPacketData(packetDataBytes)
		if err != nil {
			icaPacket, err := icacontroller.IBCMiddleware{}.UnmarshalPacketData(packetDataBytes)
			if err != nil {
				return false, "", 0, false
			}
			memo = icaPacket.(icatypes.InterchainAccountPacketData).Memo
		} else {
			memo = transferPacket.(transfertypes.FungibleTokenPacketData).Memo
		}

		err = json.Unmarshal([]byte(memo), &linkData)
		if err != nil {
			return false, "", 0, false
		}

		linkIndex, err := strconv.ParseUint(linkData.LinkIndex, 10, 64)
		if err != nil {
			return false, "", 0, false
		}

		return true, linkData.LinkID, linkIndex, linkData.IsLastPacket
	default:
		return false, "", 0, false
	}

}

func (h *ProcessProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (resp *abci.ResponseProcessProposal, err error) {
		h.Logger.Info(fmt.Sprintf("⚙️ :: Process Proposal"))

		for i, tx := range req.Txs {
			h.Logger.Info(fmt.Sprintf("⚙️:: Transaction No %v :: %v", i, tx))
		}

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}
