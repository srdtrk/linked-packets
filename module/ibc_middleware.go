package module

import (
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"

	"github.com/srdtrk/linkedpackets/keeper"
)

// IBCMiddleware implements the ICS26 callbacks for linked-packets given the
// linked-packets keeper and the underlying application.
type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper keeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddlware given the keeper and underlying application
func NewIBCMiddleware(app porttypes.IBCModule, k keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

