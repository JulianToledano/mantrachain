package unifiedAccount

import (
	"errors"

	"github.com/evmos/evmos/v20/crypto/ethsecp256k1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

func (u UnifiedAccount) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, errors.New("tx must be a SigVerifiableTx")
	}

	pubKeys, err := sigTx.GetPubKeys()
	if err != nil {
		return ctx, err
	}
	if len(pubKeys) == 0 {
		return ctx, errors.New("no pubkeys found in transaction")
	}

	for _, pk := range pubKeys {
		var cosmosAddr, ethAddr string
		switch pk.Type() {
		case "secp256k1":
			cosmosAddr, err = u.ac.BytesToString(pk.Address().Bytes())
			if err != nil {
				return ctx, err
			}

			if u.unifiedAccountExists(ctx, cosmosAddr) {
				continue
			}

			ethSecpPk := ethsecp256k1.PubKey{Key: pk.Bytes()}
			ethAddr, err = u.ac.BytesToString(ethSecpPk.Address().Bytes())
			if err != nil {
				return ctx, err
			}
		case "eth_secp256k1":
			ethAddr, err = u.ac.BytesToString(pk.Address().Bytes())
			if err != nil {
				return ctx, err
			}

			if u.unifiedAccountExists(ctx, ethAddr) {
				continue
			}

			secpPk := secp256k1.PubKey{Key: pk.Bytes()}
			cosmosAddr, err = u.ac.BytesToString(secpPk.Address().Bytes())
			if err != nil {
				return ctx, err
			}
		default:
			continue
		}

		u.mapper[cosmosAddr] = ethAddr
		u.mapper[ethAddr] = cosmosAddr
	}

	return next(ctx, tx, simulate)
}
