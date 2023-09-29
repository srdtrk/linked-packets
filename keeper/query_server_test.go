package keeper_test

import (
	"testing"

	"github.com/srdtrk/linkedpackets"
	"github.com/stretchr/testify/require"
)

func TestQueryParams(t *testing.T) {
	f := initFixture(t)
	require := require.New(t)

	resp, err := f.queryServer.Params(f.ctx, &linkedpackets.QueryParamsRequest{})
	require.NoError(err)
	require.Equal(linkedpackets.Params{}, resp.Params)
}

func TestQueryLinkEnabledChannel(t *testing.T) {
	f := initFixture(t)
	require := require.New(t)

	resp, err := f.queryServer.LinkEnabledChannel(f.ctx, &linkedpackets.QueryLinkEnabledChannelRequest{PortId: "transfer", ChannelId: "channel-0"})
	require.NoError(err)
	require.Equal(false, resp.LinkEnabled)
}
