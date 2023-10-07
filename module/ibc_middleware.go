package module

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/srdtrk/linkedpackets"
	"github.com/srdtrk/linkedpackets/keeper"
)

// IBCMiddleware implements the ICS26 callbacks for linked-packets given the
// linked-packets keeper and the underlying application.
type IBCMiddleware struct {
	app         porttypes.IBCModule
	ics4Wrapper porttypes.ICS4Wrapper

	keeper keeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddlware given the keeper and underlying application
func NewIBCMiddleware(app porttypes.IBCModule, ics4Wrapper porttypes.ICS4Wrapper, k keeper.Keeper) IBCMiddleware {
	if app == nil {
		panic(errors.New("IBCModule cannot be nil"))
	}

	if ics4Wrapper == nil {
		panic(errors.New("ICS4Wrapper cannot be nil"))
	}

	_, ok := app.(porttypes.PacketDataUnmarshaler)
	if !ok {
		panic(errors.New("IBCModule must implement PacketDataUnmarshaler"))
	}

	return IBCMiddleware{
		app:         app,
		ics4Wrapper: ics4Wrapper,
		keeper:      k,
	}
}

// OnChanOpenInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	var versionMetadata linkedpackets.Metadata

	if strings.TrimSpace(version) == "" {
		// default version
		versionMetadata = linkedpackets.Metadata{
			LinkedPacketsVersion: linkedpackets.Version,
			AppVersion:           "",
		}
	} else {
		metadata, err := linkedpackets.MetadataFromVersion(version)
		if err != nil {
			// Since it is valid for linked packets version to not be specified, the above middleware version may be for a middleware
			// lower down in the stack. Thus, if it is not a linked packets version we pass the entire version string onto the underlying
			// application.
			return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID,
				chanCap, counterparty, version)
		}
		versionMetadata = metadata
	}

	if versionMetadata.LinkedPacketsVersion != linkedpackets.Version {
		return "", errorsmod.Wrapf(linkedpackets.ErrInvalidVersion, "expected %s, got %s", linkedpackets.Version, versionMetadata.LinkedPacketsVersion)
	}

	appVersion, err := im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, versionMetadata.AppVersion)
	if err != nil {
		return "", err
	}

	versionMetadata.AppVersion = appVersion
	versionBytes, err := json.Marshal(&versionMetadata)
	if err != nil {
		return "", err
	}

	err = im.keeper.LinkEnabled.Set(ctx, collections.Join(portID, channelID))
	if err != nil {
		return "", err
	}

	// call underlying app's OnChanOpenInit callback with the appVersion
	return string(versionBytes), nil
}

// OnChanOpenTry implements the IBCMiddleware interface
// If the channel is not link enabled the underlying application version will be returned
// If the channel is link enabled we merge the underlying application version with the link version
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	versionMetadata, err := linkedpackets.MetadataFromVersion(counterpartyVersion)
	if err != nil {
		// Since it is valid for linked packets version to not be specified, the above middleware version may be for a middleware
		// lower down in the stack. Thus, if it is not a linked packets version we pass the entire version string onto the underlying
		// application.
		return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
	}

	if versionMetadata.LinkedPacketsVersion != linkedpackets.Version {
		return "", errorsmod.Wrapf(linkedpackets.ErrInvalidVersion, "expected %s, got %s", linkedpackets.Version, versionMetadata.LinkedPacketsVersion)
	}

	err = im.keeper.LinkEnabled.Set(ctx, collections.Join(portID, channelID))
	if err != nil {
		return "", err
	}

	// call underlying app's OnChanOpenTry callback with the app versions
	appVersion, err := im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, versionMetadata.AppVersion)
	if err != nil {
		return "", err
	}

	versionMetadata.AppVersion = appVersion
	versionBytes, err := json.Marshal(&versionMetadata)
	if err != nil {
		return "", err
	}

	return string(versionBytes), nil
}

// OnChanOpenAck implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	// If handshake was initialized with linked packets enabled it must complete with linked packets enabled.
	// If handshake was initialized with linked packets disabled it must complete with linked packets disabled.
	isLinkEnabled, err := im.keeper.LinkEnabled.Has(ctx, collections.Join(portID, channelID))
	if err != nil {
		return err
	}
	if isLinkEnabled {
		versionMetadata, err := linkedpackets.MetadataFromVersion(counterpartyVersion)
		if err != nil {
			return errorsmod.Wrapf(err, "failed to unmarshal ICS29 counterparty version metadata: %s", counterpartyVersion)
		}

		if versionMetadata.LinkedPacketsVersion != linkedpackets.Version {
			return errorsmod.Wrapf(linkedpackets.ErrInvalidVersion, "expected counterparty linked packets version: %s, got: %s", linkedpackets.Version, versionMetadata.LinkedPacketsVersion)
		}

		// call underlying app's OnChanOpenAck callback with the counterparty app version.
		return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, versionMetadata.AppVersion)
	}

	// call underlying app's OnChanOpenAck callback with the counterparty app version.
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// call underlying app's OnChanOpenConfirm callback.
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// call underlying app's OnChanCloseInit callback.
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// call underlying app's OnChanCloseConfirm callback.
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket implements the IBCMiddleware interface.
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	// call underlying app's OnRecvPacket callback.
	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCMiddleware interface
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	// call underlying app's OnAcknowledgementPacket callback.
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCMiddleware interface
// If fees are not enabled, this callback will default to the ibc-core packet callback
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	// call underlying app's OnTimeoutPacket callback.
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// SendPacket implements the ICS4 Wrapper interface
func (im IBCMiddleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (uint64, error) {
	isLinking, err := im.keeper.Linking.Get(ctx)
	if err != nil || !isLinking {
		return im.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	isLinkEnabled, err := im.keeper.LinkEnabled.Has(ctx, collections.Join(sourcePort, sourceChannel))
	if err != nil || !isLinkEnabled {
		return im.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	// If the channel is link enabled, then we modify the memo
	packetDataUnmarshaler, ok := im.app.(porttypes.PacketDataUnmarshaler)
	if !ok {
		panic(errors.New("IBCModule must implement PacketDataUnmarshaler"))
	}

	packetData, err := packetDataUnmarshaler.UnmarshalPacketData(data)
	if err != nil {
		return 0, err
	}

	linkId, err := im.keeper.LinkId.Get(ctx)
	if err != nil {
		return 0, err
	}

	prevPacket, err := im.keeper.PrevPacket.Get(ctx)
	if err != nil {
		prevPacket = linkedpackets.PacketIdentifier{}
	}

	var newData []byte
	var isLastPacket bool
	if transferPacket, ok := packetData.(transfertypes.FungibleTokenPacketData); ok {
		isLastPacket = strings.Contains(transferPacket.Memo, linkedpackets.LastLinkMemoKey)
		linkData := linkedpackets.LinkData{
			LinkID:         linkId,
			PrevPacket:     prevPacket,
			IsLastPacket:   isLastPacket,
			IsInitalPacket: prevPacket == linkedpackets.PacketIdentifier{},
		}

		linkDataBytes, err := json.Marshal(linkData)
		if err != nil {
			return 0, err
		}

		linkDataMemo := string(linkDataBytes)
		transferPacket.Memo = linkDataMemo

		newData = transferPacket.GetBytes()
	} else if icaPacket, ok := packetData.(icatypes.InterchainAccountPacketData); ok {
		isLastPacket = strings.Contains(icaPacket.Memo, linkedpackets.LastLinkMemoKey)

		linkData := linkedpackets.LinkData{
			LinkID:         linkId,
			PrevPacket:     prevPacket,
			IsLastPacket:   isLastPacket,
			IsInitalPacket: prevPacket == linkedpackets.PacketIdentifier{},
		}

		linkDataBytes, err := json.Marshal(linkData)
		if err != nil {
			return 0, err
		}

		linkDataMemo := string(linkDataBytes)
		icaPacket.Memo = linkDataMemo

		newData = icaPacket.GetBytes()
	} else {
		return 0, errorsmod.Wrapf(linkedpackets.ErrInvalidPacketData, "packet data type: %T", packetData)
	}

	seq, err := im.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, newData)
	if err != nil {
		return 0, err
	}

	if isLastPacket {
		err = im.keeper.PrevPacket.Remove(ctx)
		if err != nil {
			return 0, err
		}

		err = im.keeper.LinkId.Remove(ctx)
		if err != nil {
			return 0, err
		}

		err = im.keeper.Linking.Set(ctx, false)
		if err != nil {
			return 0, err
		}
	} else {
		err = im.keeper.PrevPacket.Set(ctx, linkedpackets.PacketIdentifier{
			PortId:    sourcePort,
			ChannelId: sourceChannel,
			Seq:       strconv.FormatUint(seq, 10),
		})
		if err != nil {
			return 0, err
		}
	}

	return seq, nil
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
) error {
	return im.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// GetAppVersion returns the application version of the underlying application
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return im.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
