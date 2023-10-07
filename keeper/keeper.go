package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/srdtrk/linkedpackets"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec

	// authority is the address capable of executing a MsgUpdateParams and other authority-gated message.
	// typically, this should be the x/gov module account.
	authority string

	// state management
	Schema collections.Schema
	Params collections.Item[linkedpackets.Params]
	// LinkEnabled is a KeySet of (portID, channelID) that indicates whether linked packets are enabled for a given channel.
	LinkEnabled collections.KeySet[collections.Pair[string, string]]
	Linking     collections.Item[bool]
	PrevPacket  collections.Item[linkedpackets.PacketIdentifier]
	LinkId      collections.Item[string]
	LinkIndex   collections.Map[string, uint64]
}

// NewKeeper creates a new Keeper instance
func NewKeeper(cdc codec.BinaryCodec, addressCodec address.Codec, storeService storetypes.KVStoreService, authority string) Keeper {
	if _, err := addressCodec.StringToBytes(authority); err != nil {
		panic(fmt.Errorf("invalid authority address: %w", err))
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,
		Params:       collections.NewItem(sb, linkedpackets.ParamsKey, "params", codec.CollValue[linkedpackets.Params](cdc)),
		LinkEnabled: collections.NewKeySet(
			sb, linkedpackets.LinkEnabledKey, "link_enabled", collections.PairKeyCodec(collections.StringKey, collections.StringKey),
		),
		Linking:    collections.NewItem(sb, linkedpackets.LinkingKey, "linking", collections.BoolValue),
		PrevPacket: collections.NewItem(sb, linkedpackets.PrevPacketKey, "prev_packet", codec.CollValue[linkedpackets.PacketIdentifier](cdc)),
		LinkId:     collections.NewItem(sb, linkedpackets.LinkIdKey, "link_id", collections.StringValue),
		LinkIndex:  collections.NewMap(sb, linkedpackets.LinkIndexKey, "link_index", collections.StringKey, collections.Uint64Value),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
