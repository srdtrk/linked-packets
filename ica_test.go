package linkedpackets_test

import (
	"time"

	"github.com/cosmos/gogoproto/proto"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
)

func (s *LinkedPacketsTestSuite) TestICACallbacks() {
	// Destination callbacks are not supported for ICA packets
	testCases := []struct {
		name        string
		icaMemo     string
		expSuccess  bool
	}{
		{
			"success: send ica tx with no memo",
			"",
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			icaAddr := s.SetupICATest()

			s.ExecuteICATx(icaAddr, tc.icaMemo)
		})
	}
}

func (s *LinkedPacketsTestSuite) TestICATimeoutCallbacks() {
	// ICA channels are closed after a timeout packet is executed
	testCases := []struct {
		name        string
		icaMemo     string
		expSuccess  bool
	}{
		{
			"success: send ica tx timeout with no memo",
			"",
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			icaAddr := s.SetupICATest()

			s.ExecuteICATimeout(icaAddr, tc.icaMemo)
		})
	}
}

// ExecuteICATx executes a stakingtypes.MsgDelegate on chainB by sending a packet containing the msg to chainB
func (s *LinkedPacketsTestSuite) ExecuteICATx(icaAddress, memo string) {
	timeoutTimestamp := uint64(s.chainA.GetContext().BlockTime().Add(time.Minute).UnixNano())
	icaOwner := s.chainA.SenderAccount.GetAddress().String()
	connectionID := s.path.EndpointA.ConnectionID
	// build the interchain accounts packet data
	packetData := s.buildICAMsgDelegatePacketData(icaAddress, memo)
	msg := icacontrollertypes.NewMsgSendTx(icaOwner, connectionID, timeoutTimestamp, packetData)

	res, err := s.chainA.SendMsgs(msg)
	if err != nil {
		return // we return if send packet is rejected
	}

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	err = s.path.RelayPacket(packet)
	s.Require().NoError(err)
}

// ExecuteICATx sends and times out an ICA tx
func (s *LinkedPacketsTestSuite) ExecuteICATimeout(icaAddress, memo string) {
	relativeTimeout := uint64(1)
	icaOwner := s.chainA.SenderAccount.GetAddress().String()
	connectionID := s.path.EndpointA.ConnectionID
	// build the interchain accounts packet data
	packetData := s.buildICAMsgDelegatePacketData(icaAddress, memo)
	msg := icacontrollertypes.NewMsgSendTx(icaOwner, connectionID, relativeTimeout, packetData)

	res, err := s.chainA.SendMsgs(msg)
	if err != nil {
		return // we return if send packet is rejected
	}

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	s.Require().NoError(err)

	// proof query requires up to date client
	err = s.path.EndpointA.UpdateClient()
	s.Require().NoError(err)

	err = s.path.EndpointA.TimeoutPacket(packet)
	s.Require().NoError(err)
}

// buildICAMsgDelegatePacketData builds a packetData containing a stakingtypes.MsgDelegate to be executed on chainB
func (s *LinkedPacketsTestSuite) buildICAMsgDelegatePacketData(icaAddress string, memo string) icatypes.InterchainAccountPacketData {
	// prepare a simple stakingtypes.MsgDelegate to be used as the interchain account msg executed on chainB
	validatorAddr := (sdk.ValAddress)(s.chainB.Vals.Validators[0].Address)
	msgDelegate := &stakingtypes.MsgDelegate{
		DelegatorAddress: icaAddress,
		ValidatorAddress: validatorAddr.String(),
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(5000)),
	}

	// ensure chainB is allowed to execute stakingtypes.MsgDelegate
	params := icahosttypes.NewParams(true, []string{sdk.MsgTypeURL(msgDelegate)})
	GetSimApp(s.chainB).ICAHostKeeper.SetParams(s.chainB.GetContext(), params)

	data, err := icatypes.SerializeCosmosTx(GetSimApp(s.chainA).AppCodec(), []proto.Message{msgDelegate}, icatypes.EncodingProtobuf)
	s.Require().NoError(err)

	icaPacketData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
		Memo: memo,
	}

	return icaPacketData
}
