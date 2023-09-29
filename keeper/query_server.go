package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/srdtrk/linkedpackets"
)

var _ linkedpackets.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the module QueryServer.
func NewQueryServerImpl(k Keeper) linkedpackets.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// LinkEnabledChannel defines the handler for the Query/LinkEnabledChannel RPC method.
func (qs queryServer) LinkEnabledChannel(ctx context.Context, req *linkedpackets.QueryLinkEnabledChannelRequest) (*linkedpackets.QueryLinkEnabledChannelResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	isLinkEnabled, err := qs.k.LinkEnabled.Get(ctx, collections.Join(req.PortId, req.ChannelId))
	if err != nil {
		isLinkEnabled = false
	}

	return &linkedpackets.QueryLinkEnabledChannelResponse{LinkEnabled: isLinkEnabled}, nil
}

// Params defines the handler for the Query/Params RPC method.
func (qs queryServer) Params(ctx context.Context, req *linkedpackets.QueryParamsRequest) (*linkedpackets.QueryParamsResponse, error) {
	params, err := qs.k.Params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &linkedpackets.QueryParamsResponse{Params: linkedpackets.Params{}}, nil
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &linkedpackets.QueryParamsResponse{Params: params}, nil
}
