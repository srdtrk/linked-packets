package module

import (
	"errors"

	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"

	"github.com/srdtrk/linkedpackets/keeper"
)

// IBCMiddleware implements the ICS26 callbacks for linked-packets given the
// linked-packets keeper and the underlying application.
type IBCMiddleware struct {
	app    porttypes.IBCModule
	ics4Wrapper porttypes.ICS4Wrapper
	
	keeper keeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddlware given the keeper and underlying application
func NewIBCMiddleware(app porttypes.IBCModule, ics4Wrapper porttypes.ICS4Wrapper, k keeper.Keeper) IBCMiddleware {
	if app == nil {
		panic(errors.New("IBCModule cannot be nil"))
	}

	if ics4Wrapper == nil {
		panic(errors.New("ICS4Wrapper cannot be nil"))
	}

	return IBCMiddleware{
		app:    app,
		ics4Wrapper: ics4Wrapper,
		keeper: k,
	}
}
