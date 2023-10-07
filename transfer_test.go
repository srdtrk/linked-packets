package linkedpackets_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/srdtrk/linkedpackets"

	transfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
)

func (s *LinkedPacketsTestSuite) TestTransfer() {
	testCases := []struct {
		name         string
		transferMemo string
		expSuccess   bool
	}{
		{
			"success: transfer with no memo",
			"",
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupLinkedPacketsTransferTest()

			s.ExecuteTransfer(tc.transferMemo)
		})
	}
}

func (s *LinkedPacketsTestSuite) TestLinkSuccess() {
	s.SetupLinkedPacketsTransferTest()

	s.ExecuteInitLink("mylinkid")

	packetOne := s.ExecuteTransfer("1")
	s.Require().NotNil(packetOne)
	s.Require().Equal(`{"link_id":"mylinkid","prev_packet":{},"last_packet":false,"initial_packet":true}`, packetOne.Memo)

	packetTwo := s.ExecuteTransfer("2")
	s.Require().NotNil(packetTwo)
	s.Require().Equal(`{"link_id":"mylinkid","prev_packet":{"port_id":"transfer","channel_id":"channel-0","seq":"1"},"last_packet":false,"initial_packet":false}`, packetTwo.Memo)

	packetThree := s.ExecuteTransfer("3")
	s.Require().NotNil(packetThree)
	s.Require().Equal(`{"link_id":"mylinkid","prev_packet":{"port_id":"transfer","channel_id":"channel-0","seq":"2"},"last_packet":false,"initial_packet":false}`, packetThree.Memo)

	packetFinal := s.ExecuteTransfer(linkedpackets.LastLinkMemoKey)
	s.Require().NotNil(packetFinal)
	s.Require().Equal(`{"link_id":"mylinkid","prev_packet":{"port_id":"transfer","channel_id":"channel-0","seq":"3"},"last_packet":true,"initial_packet":false}`, packetFinal.Memo)

	found, err := GetSimApp(s.chainA).LinkedPacketsKeeper.LinkId.Has(s.chainA.GetContext())
	s.Require().NoError(err)
	s.Require().False(found)

	found, err = GetSimApp(s.chainA).LinkedPacketsKeeper.PrevPacket.Has(s.chainA.GetContext())
	s.Require().NoError(err)
	s.Require().False(found)

	isLinking, err := GetSimApp(s.chainA).LinkedPacketsKeeper.Linking.Get(s.chainA.GetContext())
	s.Require().NoError(err)
	s.Require().False(isLinking)
}

func (s *LinkedPacketsTestSuite) TestTransferTimeout() {
	testCases := []struct {
		name         string
		transferMemo string
		expSuccess   bool
	}{
		{
			"success: transfer with no memo",
			"",
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupLinkedPacketsTransferTest()

			s.ExecuteTransferTimeout(tc.transferMemo)
		})

	}
}

// ExecuteTransfer executes a transfer message on chainA for ibctesting.TestCoin (100 "stake").
// It checks that the transfer is successful and that the packet is relayed to chainB.
func (s *LinkedPacketsTestSuite) ExecuteTransfer(memo string) transfertypes.FungibleTokenPacketData {
	escrowAddress := transfertypes.GetEscrowAddress(s.path.EndpointA.ChannelConfig.PortID, s.path.EndpointA.ChannelID)
	// record the balance of the escrow address before the transfer
	escrowBalance := GetSimApp(s.chainA).BankKeeper.GetBalance(s.chainA.GetContext(), escrowAddress, sdk.DefaultBondDenom)
	// record the balance of the receiving address before the transfer
	voucherDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(s.path.EndpointB.ChannelConfig.PortID, s.path.EndpointB.ChannelID, sdk.DefaultBondDenom))
	receiverBalance := GetSimApp(s.chainB).BankKeeper.GetBalance(s.chainB.GetContext(), s.chainB.SenderAccount.GetAddress(), voucherDenomTrace.IBCDenom())

	amount := ibctesting.TestCoin
	msg := transfertypes.NewMsgTransfer(
		s.path.EndpointA.ChannelConfig.PortID,
		s.path.EndpointA.ChannelID,
		amount,
		s.chainA.SenderAccount.GetAddress().String(),
		s.chainB.SenderAccount.GetAddress().String(),
		clienttypes.NewHeight(1, 100), 0, memo,
	)

	res, err := s.chainA.SendMsgs(msg)
	if err != nil {
		return transfertypes.FungibleTokenPacketData{} // we return if send packet is rejected
	}

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// relay send
	err = s.path.RelayPacket(packet)
	s.Require().NoError(err) // relay committed

	// check that the escrow address balance increased by 100
	s.Require().Equal(escrowBalance.Add(amount), GetSimApp(s.chainA).BankKeeper.GetBalance(s.chainA.GetContext(), escrowAddress, sdk.DefaultBondDenom))
	// check that the receiving address balance increased by 100
	s.Require().Equal(receiverBalance.AddAmount(sdkmath.NewInt(100)), GetSimApp(s.chainB).BankKeeper.GetBalance(s.chainB.GetContext(), s.chainB.SenderAccount.GetAddress(), voucherDenomTrace.IBCDenom()))

	packetData, err := transfer.IBCModule{}.UnmarshalPacketData(packet.Data)
	s.Require().NoError(err)
	return packetData.(transfertypes.FungibleTokenPacketData)
}

// ExecuteTransferTimeout executes a transfer message on chainA for 100 denom.
// This message is not relayed to chainB, and it times out on chainA.
func (s *LinkedPacketsTestSuite) ExecuteTransferTimeout(memo string) {
	timeoutHeight := clienttypes.GetSelfHeight(s.chainB.GetContext())
	timeoutTimestamp := uint64(s.chainB.GetContext().BlockTime().UnixNano())

	amount := ibctesting.TestCoin
	msg := transfertypes.NewMsgTransfer(
		s.path.EndpointA.ChannelConfig.PortID,
		s.path.EndpointA.ChannelID,
		amount,
		s.chainA.SenderAccount.GetAddress().String(),
		s.chainB.SenderAccount.GetAddress().String(),
		timeoutHeight, timeoutTimestamp, memo,
	)

	res, err := s.chainA.SendMsgs(msg)
	if err != nil {
		return // we return if send packet is rejected
	}

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err) // packet committed
	s.Require().NotNil(packet)

	// need to update chainA's client representing chainB to prove missing ack
	err = s.path.EndpointA.UpdateClient()
	s.Require().NoError(err)

	err = s.path.EndpointA.TimeoutPacket(packet)
	s.Require().NoError(err) // timeout committed
}
