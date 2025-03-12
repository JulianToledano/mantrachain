package unifiedAccount

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type FeeDecorator struct {
	UnifiedAccount
}

func NewUnifiedAccounFeeDecorator(unifiedAccount *UnifiedAccount) FeeDecorator {
	return FeeDecorator{UnifiedAccount: *unifiedAccount}
}

func (f FeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if simulate {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if err := f.shareFees(ctx, feeTx); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (f FeeDecorator) shareFees(ctx sdk.Context, feeTx sdk.FeeTx) error {
	if feeTx.FeeGranter() != nil {
		return nil
	}

	fee := feeTx.GetFee()
	feePayer := feeTx.FeePayer()

	if f.bk.GetBalance(ctx, feePayer, fee[0].Denom).IsGTE(fee[0]) {
		return nil
	}

	feePayerAddr, err := f.ac.BytesToString(feePayer)
	if err != nil {
		return err
	}

	pairAddr, err := f.getPairAddress(feePayerAddr)
	if err != nil {
		return err
	}

	return f.bk.SendCoins(ctx, pairAddr, feePayer, fee)
}
