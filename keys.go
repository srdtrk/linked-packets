package linkedpackets

import "cosmossdk.io/collections"

const (
	ModuleName = "linkedpackets"

	// StoreKey is the string store representation
	StoreKey = ModuleName
	
	Version    = "ics29-1"
)

var (
	ParamsKey      = collections.NewPrefix(0)
	LinkEnabledKey = collections.NewPrefix(1)
)
