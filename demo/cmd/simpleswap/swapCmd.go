package main

import (
	"context"
	"crypto/sha256"
	"decred.org/dcrdex/client/asset"
	"decred.org/dcrdex/dex"
	"decred.org/dcrdex/dex/encode"
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/decred/dcrd/dcrutil/v3"
	"time"
)

type swapCmd struct {
	fromCoin   string
	fromAmount float64
	toCoin     string
	toAmount   float64
}

func (c *swapCmd) extractRole(party1WM, party2WM *walletMatcher) (fromDCR, toDCR, fromBTC, toBTC asset.Wallet, amountBTC btcutil.Amount, amountDCR dcrutil.Amount, err error) {
	if c.fromCoin == "dcr" && c.toCoin == "btc" {
		fromDCR, toDCR, fromBTC, toBTC = party1WM.dcrWallet, party2WM.dcrWallet, party2WM.btcWallet, party1WM.btcWallet
		amountDCR, _ = dcrutil.NewAmount(c.fromAmount)
		amountBTC, _ = btcutil.NewAmount(c.toAmount)
		return
	}
	if c.fromCoin == "btc" && c.toCoin == "dcr" {
		fromDCR, toDCR, fromBTC, toBTC = party2WM.dcrWallet, party1WM.dcrWallet, party1WM.btcWallet, party2WM.btcWallet
		amountDCR, _ = dcrutil.NewAmount(c.toAmount)
		amountBTC, _ = btcutil.NewAmount(c.fromAmount)
		return
	}
	err = fmt.Errorf("Invalid input")
	return
}

func getDexAsset(symbol string) dex.Asset {
	if symbol == "dcr" {
		return dex.Asset{
			ID:           0,
			Symbol:       symbol,
			LotSize:      100000000,
			RateStep:     100000000,
			MaxFeeRate:   10,
			SwapSize:     0,
			SwapSizeBase: 0,
			SwapConf:     4,
		}
	}
	if symbol == "btc" {
		return dex.Asset{
			ID:           0,
			Symbol:       symbol,
			LotSize:      100000,
			RateStep:     100000,
			MaxFeeRate:   100,
			SwapSize:     0,
			SwapSizeBase: 0,
			SwapConf:     1,
		}
	}
	return dex.Asset{}
}

func (c *swapCmd) swap(symbol string, fromW, toW asset.Wallet, amount uint64, secretHash []byte) (redeemScripts []dex.Bytes, receipts []asset.Receipt, feePaid uint64, err error) {
	dexAsset := getDexAsset(symbol)
	order := asset.Order{
		Value:        amount,
		MaxSwapCount: amount / dexAsset.LotSize,
		DEXConfig:    &dexAsset,
		Immediate:    true,
	}
	balance,_ := fromW.Balance()
	fmt.Printf("Balance(Available: %d, Immature: %d, Locked: %d) \n", balance.Available, balance.Immature, balance.Locked)
	coins, redeemScript, err := fromW.FundOrder(&order)
	fmt.Println(err)
	if err != nil {
		return nil, nil, 0, err
	}
	toAddr, err := toW.Address()
	if err != nil {
		return nil, nil, 0, err
	}
	contract := asset.Contract{
		Address:    toAddr,
		Value:      amount,
		SecretHash: secretHash[:],
		LockTime:   uint64(time.Now().Add(time.Hour * 24).UTC().Unix()),
	}
	swap := asset.Swaps{
		Inputs:     coins,
		Contracts:  []*asset.Contract{&contract},
		FeeRate:    0,
		LockChange: false,
	}
	receipts,_,feePaid, err = fromW.Swap(&swap)

	return redeemScript, receipts, feePaid, err
}

func (c *swapCmd) runCommand(ctx context.Context, party1WM, party2WM *walletMatcher) error {
	fromDCR, toDCR, fromBTC, toBTC, amountBTC, amountDCR, err := c.extractRole(party1WM, party2WM)
	if err != nil {
		return err
	}
	secret := encode.RandomBytes(32)
	secretHash := sha256.Sum256(secret)
	dcrRS, dcrReceipts, dcrFee, err := c.swap("dcr", fromDCR, toDCR, uint64(amountDCR), secretHash[:])
	if err != nil {
		return err
	}
	fmt.Printf("DCR swap info: %v, %v, %v \n", dcrRS, dcrReceipts, dcrFee)
	btcRS, btcReceipts, btcFee, err := c.swap("btc", fromBTC, toBTC, uint64(amountBTC), secretHash[:])
	if err != nil {
		return err
	}
	fmt.Printf("BTC swap info: %v, %v, %v \n", btcRS, btcReceipts, btcFee)
	return nil
}
