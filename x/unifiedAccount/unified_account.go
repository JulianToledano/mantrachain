package unifiedAccount

import (
	"errors"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

/*
 UnifiedAccount module provides functionality for managing accounts that can represent multiple address types.
 It includes features for handling transaction fees and bank sends operations.
 The module defines the UnifiedAccount structure, which contains a bankkeeper for managing balances,
 a mapper for converting between secp256k1 and eth_secp256k1 and a codec for address encoding and decoding.

 Key components of the module include:
 - FeeDecorator: A decorator that processes transaction fees and attempts to share fees with a secondary address if the
   fee payer's balance is insufficient.
 - AnteHandle: A function that manages the mapping between secp256k1 and eth_secp256k1 addresses during transaction processing.
 - BankHooks: Hooks that are triggered before sending coins, ensuring that the sender has sufficient balance and
   covering any deficits from a secondary address.

 PROS:
 - Users don't need to manually transfer funds between accounts.
 - Transactions can succeed even if the used account has insufficient funds.
 - Simple to understand and maintain.

 CONS:
 - Could lead to unexpected fund movements that users might not notice immediately.
 - Each automatic transfer requires an additional transaction.
 - Users should pay more gas fees due to the extra transfer, but this is not actually being calculated or paid.
 - Current implementation silently fails in error cases.
 - Users need to understand the relationship between accounts.
 - This is automatically enabled, user should have the right to opt out.
 - This implementation only works with the bank send msg. In theory a user should be able to use funds for any kind of msg. (What about a future swap message for example?)

 TODO IMPROVEMENTS:
 - Instead of a map in memory we could save this in state (it has its perks and cons).
 - Instead of a map a cache should be used for storing the pair addresses.
 - We should emit custom events for the automatic transfers or errors.

 CONCLUSION:
 IMHO this is a client side problem, not a protocol problem. Really feel this solution is not a good fit for the
 protocol, as most modules use addresses to store state information. If an ante handler solution is applied, I believe
 is not a good approach as txs should either succeed or fail as whole. Moving funds in the AnteHandler introduces state
 changes before the transaction are fully validated and executed.

 The idea should be that any pair of addresses should be able to be used for any kind of msg. This means that there must
 be an "interceptor" that can check the amount field for all messages to find out the necessary funds an address is
 missing.

 There may be more suitable approaches as a middleware between ante handler and tx execution, although an ante handler
 for fee sharing is still needed.
*/

// UnifiedAccount represents a structure that manages the mapping between different address types
// and handles the bank operations for those addresses. It contains a bankkeeper for managing balances,
// a mapper for address conversions, and a codec for address encoding and decoding.
type UnifiedAccount struct {
	bk     bankkeeper.Keeper // Keeper for managing bank operations
	mapper map[string]string // Maps addresses between different types (e.g., secp256k1 and eth_secp256k1)
	ac     address.Codec     // Codec for address encoding and decoding
}

func NewUnifiedAccount(bk bankkeeper.Keeper, ac address.Codec) UnifiedAccount {
	return UnifiedAccount{
		bk:     bk,
		mapper: make(map[string]string),
		ac:     ac,
	}
}

// getPairAddress retrieves the corresponding unified address for a given address as a byte slice (AccAddress).
func (u UnifiedAccount) getPairAddress(addr string) ([]byte, error) {
	pairAddr, ok := u.mapper[addr]
	if !ok {
		return nil, errors.New("no unified address")
	}

	return u.ac.StringToBytes(pairAddr)
}

// unifiedAccountExists checks if given address has been mapped to a unified address.
func (u UnifiedAccount) unifiedAccountExists(_ sdk.Context, addr string) bool {
	_, ok := u.mapper[addr]
	return ok
}
