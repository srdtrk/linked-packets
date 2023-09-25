package keeper

import (
	"context"

	"github.com/srdtrk/linkedpackets"
)

// InitGenesis initializes the module state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *linkedpackets.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}

	return nil
}

// ExportGenesis exports the module state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) (*linkedpackets.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	var counters []linkedpackets.Counter

	return &linkedpackets.GenesisState{
		Params:   params,
		Counters: counters,
	}, nil
}
