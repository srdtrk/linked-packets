package keeper_test

import (
	"fmt"
	"testing"

	"github.com/srdtrk/linkedpackets"
	"github.com/stretchr/testify/require"
)

func TestUpdateParams(t *testing.T) {
	f := initFixture(t)
	require := require.New(t)

	testCases := []struct {
		name         string
		request      *linkedpackets.MsgUpdateParams
		expectErrMsg string
	}{
		{
			name: "set invalid authority (not an address)",
			request: &linkedpackets.MsgUpdateParams{
				Authority: "foo",
			},
			expectErrMsg: "invalid authority address",
		},
		{
			name: "set invalid authority (not defined authority)",
			request: &linkedpackets.MsgUpdateParams{
				Authority: f.addrs[1].String(),
			},
			expectErrMsg: fmt.Sprintf("unauthorized, authority does not match the module's authority: got %s, want %s", f.addrs[1].String(), f.k.GetAuthority()),
		},
		{
			name: "set valid params",
			request: &linkedpackets.MsgUpdateParams{
				Authority: f.k.GetAuthority(),
				Params:    linkedpackets.Params{},
			},
			expectErrMsg: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.UpdateParams(f.ctx, tc.request)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestInitLink(t *testing.T) {
	f := initFixture(t)
	require := require.New(t)

	testCases := []struct {
		name         string
		request      *linkedpackets.MsgInitLink
		expectErrMsg string
	}{
		{
			name: "set invalid sender (not an address)",
			request: &linkedpackets.MsgInitLink{
				Sender: "foo",
			},
			expectErrMsg: "invalid sender address",
		},
		{
			name: "set valid sender",
			request: &linkedpackets.MsgInitLink{
				Sender: "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5",
			},
			expectErrMsg: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.InitLink(f.ctx, tc.request)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)
				val, err := f.k.Linking.Get(f.ctx)
				require.NoError(err)
				require.Equal(true, val)
			}
		})
	}
}

func TestStopLink(t *testing.T) {
	f := initFixture(t)
	require := require.New(t)

	testCases := []struct {
		name         string
		request      *linkedpackets.MsgStopLink
		expectErrMsg string
	}{
		{
			name: "set invalid sender (not an address)",
			request: &linkedpackets.MsgStopLink{
				Sender: "foo",
			},
			expectErrMsg: "invalid sender address",
		},
		{
			name: "set valid sender",
			request: &linkedpackets.MsgStopLink{
				Sender: "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5",
			},
			expectErrMsg: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.StopLink(f.ctx, tc.request)
			if tc.expectErrMsg != "" {
				require.Error(err)
				require.ErrorContains(err, tc.expectErrMsg)
			} else {
				require.NoError(err)
				val, err := f.k.Linking.Get(f.ctx)
				require.NoError(err)
				require.Equal(false, val)
			}
		})
	}
}
