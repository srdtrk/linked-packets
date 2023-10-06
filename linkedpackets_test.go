package linkedpackets_test

import (
	"encoding/json"
	"errors"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	feetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	ibcmock "github.com/cosmos/ibc-go/v8/testing/mock"

	"github.com/srdtrk/linkedpackets/simapp"
)

func init() {
	ibctesting.DefaultTestingAppInit = SetupTestingApp
}

// SetupTestingApp provides the duplicated simapp which is specific to the callbacks module on chain creation.
func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	app := simapp.NewSimApp(log.NewNopLogger(), db, nil, true, simtestutil.EmptyAppOptions{})
	return app, app.DefaultGenesis()
}

// GetSimApp returns the duplicated SimApp from within the callbacks directory.
// This must be used instead of chain.GetSimApp() for tests within this directory.
func GetSimApp(chain *ibctesting.TestChain) *simapp.SimApp {
	app, ok := chain.App.(*simapp.SimApp)
	if !ok {
		panic(errors.New("chain is not a simapp.SimApp"))
	}
	return app
}

// LinkedPacketsTestSuite defines the needed instances and methods to test callbacks
type LinkedPacketsTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	path *ibctesting.Path
}

// setupChains sets up a coordinator with 2 test chains.
func (s *LinkedPacketsTestSuite) setupChains() {
	s.coordinator = ibctesting.NewCoordinator(s.T(), 2)
	s.chainA = s.coordinator.GetChain(ibctesting.GetChainID(1))
	s.chainB = s.coordinator.GetChain(ibctesting.GetChainID(2))
	s.path = ibctesting.NewPath(s.chainA, s.chainB)
}

// SetupTransferTest sets up a transfer channel between chainA and chainB
func (s *LinkedPacketsTestSuite) SetupTransferTest() {
	s.setupChains()

	s.path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	s.path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	s.path.EndpointA.ChannelConfig.Version = transfertypes.Version
	s.path.EndpointB.ChannelConfig.Version = transfertypes.Version

	s.coordinator.Setup(s.path)
}

// SetupFeeTransferTest sets up a fee middleware enabled transfer channel between chainA and chainB
func (s *LinkedPacketsTestSuite) SetupFeeTransferTest() {
	s.setupChains()

	feeTransferVersion := string(feetypes.ModuleCdc.MustMarshalJSON(&feetypes.Metadata{FeeVersion: feetypes.Version, AppVersion: transfertypes.Version}))
	s.path.EndpointA.ChannelConfig.Version = feeTransferVersion
	s.path.EndpointB.ChannelConfig.Version = feeTransferVersion
	s.path.EndpointA.ChannelConfig.PortID = transfertypes.PortID
	s.path.EndpointB.ChannelConfig.PortID = transfertypes.PortID

	s.coordinator.Setup(s.path)
}

func (s *LinkedPacketsTestSuite) SetupMockFeeTest() {
	s.setupChains()

	mockFeeVersion := string(feetypes.ModuleCdc.MustMarshalJSON(&feetypes.Metadata{FeeVersion: feetypes.Version, AppVersion: ibcmock.Version}))
	s.path.EndpointA.ChannelConfig.Version = mockFeeVersion
	s.path.EndpointB.ChannelConfig.Version = mockFeeVersion
	s.path.EndpointA.ChannelConfig.PortID = ibctesting.MockFeePort
	s.path.EndpointB.ChannelConfig.PortID = ibctesting.MockFeePort
}

// SetupICATest sets up an interchain accounts channel between chainA (controller) and chainB (host).
// It funds and returns the interchain account address owned by chainA's SenderAccount.
func (s *LinkedPacketsTestSuite) SetupICATest() string {
	s.setupChains()
	s.coordinator.SetupConnections(s.path)

	icaOwner := s.chainA.SenderAccount.GetAddress().String()
	// ICAVersion defines a interchain accounts version string
	icaVersion := icatypes.NewDefaultMetadataString(s.path.EndpointA.ConnectionID, s.path.EndpointB.ConnectionID)
	icaControllerPortID, err := icatypes.NewControllerPortID(icaOwner)
	s.Require().NoError(err)

	s.path.SetChannelOrdered()
	s.path.EndpointA.ChannelConfig.PortID = icaControllerPortID
	s.path.EndpointB.ChannelConfig.PortID = icatypes.HostPortID
	s.path.EndpointA.ChannelConfig.Version = icaVersion
	s.path.EndpointB.ChannelConfig.Version = icaVersion

	s.RegisterInterchainAccount(icaOwner)
	// open chan init must be skipped. So we cannot use .CreateChannels()
	err = s.path.EndpointB.ChanOpenTry()
	s.Require().NoError(err)
	err = s.path.EndpointA.ChanOpenAck()
	s.Require().NoError(err)
	err = s.path.EndpointB.ChanOpenConfirm()
	s.Require().NoError(err)

	interchainAccountAddr, found := GetSimApp(s.chainB).ICAHostKeeper.GetInterchainAccountAddress(s.chainB.GetContext(), s.path.EndpointA.ConnectionID, s.path.EndpointA.ChannelConfig.PortID)
	s.Require().True(found)

	// fund the interchain account on chainB
	msgBankSend := &banktypes.MsgSend{
		FromAddress: s.chainB.SenderAccount.GetAddress().String(),
		ToAddress:   interchainAccountAddr,
		Amount:      sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000))),
	}
	res, err := s.chainB.SendMsgs(msgBankSend)
	s.Require().NotEmpty(res)
	s.Require().NoError(err)

	return interchainAccountAddr
}

// RegisterInterchainAccount submits a MsgRegisterInterchainAccount and updates the controller endpoint with the
// channel created.
func (s *LinkedPacketsTestSuite) RegisterInterchainAccount(owner string) {
	msgRegister := icacontrollertypes.NewMsgRegisterInterchainAccount(s.path.EndpointA.ConnectionID, owner, s.path.EndpointA.ChannelConfig.Version)

	res, err := s.chainA.SendMsgs(msgRegister)
	s.Require().NotEmpty(res)
	s.Require().NoError(err)

	channelID, err := ibctesting.ParseChannelIDFromEvents(res.Events)
	s.Require().NoError(err)

	s.path.EndpointA.ChannelID = channelID
}

func TestIBCLinkedPacketsTestSuite(t *testing.T) {
	suite.Run(t, new(LinkedPacketsTestSuite))
}
