package linkedpackets

import errorsmod "cosmossdk.io/errors"

var (
	// ErrInvalidVersion error if the channel version is invalid
	ErrInvalidVersion = errorsmod.Register(ModuleName, 2, "invalid linked packets middleware version")
	// ErrDuplicateAddress error if there is a duplicate address
	ErrDuplicateAddress  = errorsmod.Register(ModuleName, 3, "duplicate address")
	ErrInvalidPacketData = errorsmod.Register(ModuleName, 4, "invalid packet data")
)
