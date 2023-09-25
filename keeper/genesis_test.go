package keeper_test

import (
	"testing"

	"github.com/srdtrk/linkedpackets"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	fixture := initFixture(t)

	data := &linkedpackets.GenesisState{
		Counters: []linkedpackets.Counter{
			{
				Address: fixture.addrs[0].String(),
				Count:   5,
			},
		},
		Params: linkedpackets.DefaultParams(),
	}
	err := fixture.k.InitGenesis(fixture.ctx, data)
	require.NoError(t, err)

	params, err := fixture.k.Params.Get(fixture.ctx)
	require.NoError(t, err)

	require.Equal(t, linkedpackets.DefaultParams(), params)
}

func TestExportGenesis(t *testing.T) {
	fixture := initFixture(t)

	_, err := fixture.msgServer.InitLink(fixture.ctx, &linkedpackets.MsgInitLink{
		Sender: fixture.addrs[0].String(),
	})
	require.NoError(t, err)

	_, err = fixture.k.ExportGenesis(fixture.ctx)
	require.NoError(t, err)

	// require.Equal(t, linkedpackets.DefaultParams(), out.Params)
	// require.Equal(t, uint64(0), out.Counters[0].Count)
}
