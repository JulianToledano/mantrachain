package ante

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/evmos/evmos/v20/crypto/ethsecp256k1"

	"github.com/evmos/evmos/v20/app/ante"
)

// SigVerificationGasConsumer is a copy of evmos SigVerificationGasConsumer that enables cosmos secp256k1.
// https://github.com/evmos/evmos/blob/main/app/ante/sigverify.go#L37-L63
func SigVerificationGasConsumer(
	meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params,
) error {
	pubkey := sig.PubKey
	switch pubkey := pubkey.(type) {

	case *ethsecp256k1.PubKey:
		// Ethereum keys
		meter.ConsumeGas(ante.Secp256k1VerifyCost, "ante verify: eth_secp256k1")
		return nil
	case *ed25519.PubKey:
		// Validator keys
		meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
		return errorsmod.Wrap(errortypes.ErrInvalidPubKey, "ED25519 public keys are unsupported")
	case *secp256k1.PubKey:
		// cosmos secp256k1
		meter.ConsumeGas(params.SigVerifyCostSecp256k1, "ante verify: secp256k1")
		return nil
	case multisig.PubKey:
		// Multisig keys
		multisignature, ok := sig.Data.(*signing.MultiSignatureData)
		if !ok {
			return fmt.Errorf("expected %T, got, %T", &signing.MultiSignatureData{}, sig.Data)
		}
		return ante.ConsumeMultisignatureVerificationGas(meter, multisignature, pubkey, params, sig.Sequence)

	default:
		return errorsmod.Wrapf(errortypes.ErrInvalidPubKey, "unrecognized/unsupported public key type: %T", pubkey)
	}
}
