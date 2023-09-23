package linkedpackets

import "cosmossdk.io/collections"

const ModuleName = "linkedpackets"

var (
	ParamsKey  = collections.NewPrefix(0)
	CounterKey = collections.NewPrefix(1)
)
