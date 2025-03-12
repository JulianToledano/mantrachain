package unifiedAccount

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ types.BankHooks = UnifiedAccount{}

func (u UnifiedAccount) TrackBeforeSend(ctx context.Context, from, _ sdk.AccAddress, amount sdk.Coins) {
	fromBalance := u.bk.GetBalance(ctx, from, amount[0].Denom)
	if fromBalance.IsGTE(amount[0]) {
		return
	}

	secondAddr, ok := u.mapper[from.String()]
	if !ok {
		return
	}

	neededBalance := amount[0].Sub(fromBalance)
	addr, err := u.ac.StringToBytes(secondAddr)
	if err != nil {
		return
	}

	err = u.bk.SendCoins(ctx, addr, from, sdk.Coins{neededBalance})
	if err != nil {
		return
	}
}

func (u UnifiedAccount) BlockBeforeSend(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	return nil
}
