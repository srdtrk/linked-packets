package keeper

import (
	"context"
	"errors"
	"fmt"

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

// Counter defines the handler for the Query/Counter RPC method.
func (qs queryServer) Counter(ctx context.Context, req *linkedpackets.QueryCounterRequest) (*linkedpackets.QueryCounterResponse, error) {
	if _, err := qs.k.addressCodec.StringToBytes(req.Address); err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	counter, err := qs.k.Counter.Get(ctx, req.Address)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return &linkedpackets.QueryCounterResponse{Counter: 0}, nil
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &linkedpackets.QueryCounterResponse{Counter: counter}, nil
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
