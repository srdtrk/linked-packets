package linkedpackets

import "cosmossdk.io/collections"

const (
	ModuleName = "linkedpackets"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	Version = "ics29-1"
)

var (
	ParamsKey      = collections.NewPrefix(0)
	LinkEnabledKey = collections.NewPrefix(1)
	LinkingKey     = collections.NewPrefix(2)
	PrevPacketKey  = collections.NewPrefix(3)
	LinkIdKey      = collections.NewPrefix(4)
	LinkIndexKey   = collections.NewPrefix(5)
)
