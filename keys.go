package linkedpackets

import "cosmossdk.io/collections"

const (
	ModuleName = "linkedpackets"
	Version = "ics29-1"
)

var (
	ParamsKey  = collections.NewPrefix(0)
	CounterKey = collections.NewPrefix(1)
)
