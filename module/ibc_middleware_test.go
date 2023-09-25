package module_test

import (
	"errors"
	"testing"

	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"
	ibcmock "github.com/cosmos/ibc-go/v8/testing/mock"

	"github.com/stretchr/testify/require"

	"github.com/srdtrk/linkedpackets/keeper"
	"github.com/srdtrk/linkedpackets/module"
)

func TestNewIBCMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		instantiateFn func()
		expError      error
	}{
		{
			"success",
			func() {
				_ = module.NewIBCMiddleware(ibcmock.IBCModule{}, channelkeeper.Keeper{}, keeper.Keeper{})
			},
			nil,
		},
		{
			"failure: app is nil",
			func() {
				_ = module.NewIBCMiddleware(nil, channelkeeper.Keeper{}, keeper.Keeper{})
			},
			errors.New("IBCModule cannot be nil"),
		},
		{
			"failure: ics4Wrapper is nil",
			func() {
				_ = module.NewIBCMiddleware(ibcmock.IBCModule{}, nil, keeper.Keeper{})
			},
			errors.New("ICS4Wrapper cannot be nil"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			expPass := tc.expError == nil
			if expPass {
				require.NotPanics(t, tc.instantiateFn, "unexpected panic: NewIBCMiddleware")
			} else {
				require.PanicsWithError(t, tc.expError.Error(), tc.instantiateFn, "expected panic with error: ", tc.expError.Error())
			}
		})
	}
}
