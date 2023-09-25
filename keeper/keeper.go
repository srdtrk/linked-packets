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
	// LinkEnabled is a map of (portID, channelID) -> bool that indicates whether linked packets are enabled for a given channel.
	LinkEnabled collections.Map[collections.Pair[string, string], bool]
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
		LinkEnabled: collections.NewMap(
			sb, linkedpackets.LinkEnabledKey, "link_enabled", collections.PairKeyCodec(collections.StringKey, collections.StringKey), collections.BoolValue,
		),
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
