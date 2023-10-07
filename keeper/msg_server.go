package keeper

import (
	"context"
	"fmt"
	"strings"

	"github.com/srdtrk/linkedpackets"
)

type msgServer struct {
	k Keeper
}

var _ linkedpackets.MsgServer = (*msgServer)(nil)

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) linkedpackets.MsgServer {
	return &msgServer{k: keeper}
}

// InitLink defines the handler for the MsgInitLink message.
func (ms msgServer) InitLink(ctx context.Context, msg *linkedpackets.MsgInitLink) (*linkedpackets.MsgInitLinkResponse, error) {
	if _, err := ms.k.addressCodec.StringToBytes(msg.Sender); err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	err := ms.k.Linking.Set(ctx, true)
	if err != nil {
		return nil, err
	}

	err = ms.k.LinkId.Set(ctx, msg.LinkId)

	return &linkedpackets.MsgInitLinkResponse{}, nil
}

// StopLink defines the handler for the MsgStopLink message.
func (ms msgServer) StopLink(ctx context.Context, msg *linkedpackets.MsgStopLink) (*linkedpackets.MsgStopLinkResponse, error) {
	if _, err := ms.k.addressCodec.StringToBytes(msg.Sender); err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	err := ms.k.Linking.Set(ctx, false)
	if err != nil {
		return nil, err
	}

	err = ms.k.PrevPacket.Remove(ctx)
	if err != nil {
		return nil, err
	}

	err = ms.k.LinkId.Remove(ctx)
	if err != nil {
		return nil, err
	}

	return &linkedpackets.MsgStopLinkResponse{}, nil
}

// UpdateParams params is defining the handler for the MsgUpdateParams message.
func (ms msgServer) UpdateParams(ctx context.Context, msg *linkedpackets.MsgUpdateParams) (*linkedpackets.MsgUpdateParamsResponse, error) {
	if _, err := ms.k.addressCodec.StringToBytes(msg.Authority); err != nil {
		return nil, fmt.Errorf("invalid authority address: %w", err)
	}

	if authority := ms.k.GetAuthority(); !strings.EqualFold(msg.Authority, authority) {
		return nil, fmt.Errorf("unauthorized, authority does not match the module's authority: got %s, want %s", msg.Authority, authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := ms.k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &linkedpackets.MsgUpdateParamsResponse{}, nil
}
