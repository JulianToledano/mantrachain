package unifiedAccount

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type FeeDecorator struct {
	UnifiedAccount
}

func NewUnifiedAccountFeeDecorator(unifiedAccount *UnifiedAccount) FeeDecorator {
	return FeeDecorator{UnifiedAccount: *unifiedAccount}
}

// AnteHandle processes the transaction fees for a given transaction.
// Checks if sender has enough balance to pay for the fees, if not, it will try to share the fees with the second address.
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

// shareFees attempts to share transaction fees with a secondary address if the fee payer does not have sufficient balance.
// If the fee payer's balance is insufficient, it retrieves the pair address and attempts to send the required fees from the secondary address.
func (f FeeDecorator) shareFees(ctx sdk.Context, feeTx sdk.FeeTx) error {
	if feeTx.FeeGranter() != nil {
		return nil
	}

	fee := feeTx.GetFee()
	if len(fee) == 0 {
		return nil
	}

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

	// TODO: Not a fan of this, doing sends in the ante handler is not a good idea.
	return f.bk.SendCoins(ctx, pairAddr, feePayer, fee)
}
