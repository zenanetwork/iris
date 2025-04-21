package zena

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	jsoniter "github.com/json-iterator/go"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/zenanetwork/iris/helper"
	hmTypes "github.com/zenanetwork/iris/types"
	"github.com/zenanetwork/iris/zena/client/rest"
	"github.com/zenanetwork/iris/zena/types"
)

// ResponseWithHeight defines a response object type that wraps an original
// response with a height.
type ResponseWithHeight struct {
	Height string              `json:"height"`
	Result jsoniter.RawMessage `json:"result"`
}

func BeginBlocker(ctx sdk.Context, _ abci.RequestBeginBlock, k Keeper) {
	if ctx.BlockHeight() == helper.GetSpanOverrideHeight() {
		k.Logger(ctx).Info("overriding span BeginBlocker", "height", ctx.BlockHeight())

		j, ok := rest.SPAN_OVERRIDES[helper.GenesisDoc.ChainID]
		if !ok {
			k.Logger(ctx).Info("No Override span found")
			return
		}

		var spans []*types.ResponseWithHeight

		if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(j, &spans); err != nil {
			k.Logger(ctx).Error("Error Unmarshal spans", "error", err)
			panic(err)
		}

		for _, span := range spans {
			k.Logger(ctx).Info("overriding span", "height", span.Height, "span", span)

			var irisSpan hmTypes.Span
			if err := jsoniter.ConfigFastest.Unmarshal(span.Result, &irisSpan); err != nil {
				k.Logger(ctx).Error("Error Unmarshal irisSpan", "error", err)
				panic(err)
			}

			if err := k.AddNewRawSpan(ctx, irisSpan); err != nil {
				k.Logger(ctx).Error("Error AddNewRawSpan", "error", err)
				panic(err)
			}

			k.UpdateLastSpan(ctx, irisSpan.ID)
		}
	}
}
