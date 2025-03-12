package unifiedAccount

import (
	"errors"

	"cosmossdk.io/core/address"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type UnifiedAccount struct {
	bk     bankkeeper.Keeper
	mapper map[string]string // TODO: make sense to be in state?
	ac     address.Codec
}

func NewUnifiedAccount(bk bankkeeper.Keeper, ac address.Codec) UnifiedAccount {
	return UnifiedAccount{
		bk:     bk,
		mapper: make(map[string]string),
		ac:     ac,
	}
}

func (u UnifiedAccount) getPairAddress(addr string) ([]byte, error) {
	pairAddr, ok := u.mapper[addr]
	if !ok {
		return nil, errors.New("no unified address")
	}

	return u.ac.StringToBytes(pairAddr)
}

func (u UnifiedAccount) unifiedAccountExists(_ sdk.Context, addr string) bool {
	_, ok := u.mapper[addr]
	return ok
}
