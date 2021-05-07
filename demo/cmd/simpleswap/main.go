package main

import (
	"context"
	"crypto/sha256"
	"decred.org/dcrdex/client/asset"
	_ "decred.org/dcrdex/client/asset/btc"
	_ "decred.org/dcrdex/client/asset/dcr"
	"decred.org/dcrdex/dex"
	"decred.org/dcrdex/dex/encode"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type walletMatcher struct {
	btcWallet asset.Wallet
	dcrWallet asset.Wallet
}

type command interface {
	runCommand(context.Context, *walletMatcher) error
}

var (
	flagset     = flag.NewFlagSet("", flag.ExitOnError)
	confFlag    = flagset.String("conf", "demo/config.json", "path to wallet connection config file")
	testnetFlag = flagset.Bool("testnet", false, "use testnet network")
)

func init() {
	flagset.Usage = func() {
		fmt.Println("Usage: simpleswap [flags] cmd [cmd args]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  initiate <coin type: btc or dcr> <amount>")
		fmt.Println("  participate <initiator address> <amount> <secret hash>")
		fmt.Println("  redeem <contract> <contract transaction> <secret>")
		fmt.Println("  refund <contract> <contract transaction>")
		fmt.Println("  extractsecret <redemption transaction> <secret hash>")
		fmt.Println("  auditcontract <contract> <contract transaction>")
		fmt.Println()
		fmt.Println("Flags:")
		flagset.PrintDefaults()
	}
}

func main() {
	err, showUse := _main()
	if showUse {
		flagset.Usage()
		return
	}
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
		return
	}
	fmt.Println("Success!!!")
	os.Exit(0)
}

func _main() (error, bool) {
	flagset.Parse(os.Args[1:])
	args := flagset.Args()
	if len(args) == 0 {
		return nil, true
	}
	cmdArgs := 0
	switch args[0] {
	case "initiate":
		cmdArgs = 2
	case "participate":
		cmdArgs = 3
	case "redeem":
		cmdArgs = 3
	case "refund":
		cmdArgs = 2
	case "extractsecret":
		cmdArgs = 2
	case "auditcontract":
		cmdArgs = 2
	default:
		return fmt.Errorf("unknown command %v", args[0]), true
	}
	nArgs := checkCmdArgLength(args[1:], cmdArgs)
	flagset.Parse(args[1+nArgs:])
	if nArgs < cmdArgs {
		return fmt.Errorf("%s: too few arguments", args[0]), true
	}
	if flagset.NArg() != 0 {
		return fmt.Errorf("unexpected argument: %s", flagset.Arg(0)), true
	}
	err := flagset.Parse(os.Args[1:])
	wm, err := initWallet(*confFlag)
	if err != nil {
		return err, false
	}

	var cmd command
	switch args[0] {
	case "initiate":
		amount, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return fmt.Errorf("Amount must be number: ", err), false
		}
		cmd = &initiateCmd{
			coin:   args[1],
			amount: amount,
		}
	}
	err = cmd.runCommand(context.Background(), wm)
	return err, false
}

func checkCmdArgLength(args []string, required int) (nArgs int) {
	if len(args) < required {
		return 0
	}
	for i, arg := range args[:required] {
		if len(arg) != 1 && strings.HasPrefix(arg, "-") {
			return i
		}
	}
	return required
}

func initWallet(confFilePath string) (*walletMatcher, error) {
	data, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}
	confInfo := make(map[string]map[string]string)
	err = json.Unmarshal(data, &confInfo)
	if err != nil {
		return nil, err
	}
	btcSetting, ok := confInfo["btc"]
	if !ok {
		return nil, fmt.Errorf("Configure for btc wallet is required")
	}
	// setup btc wallet
	btcConf := asset.WalletConfig{
		Settings: btcSetting,
		TipChange: func(e error) {

		},
	}
	btcWallet, err := asset.Setup(0, &btcConf, dex.NewLogger("BTC", dex.LevelTrace, os.Stdout), dex.Testnet)
	if err != nil {
		return nil, err
	}
	_, err = btcWallet.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	if err = btcWallet.Unlock(btcSetting["walletpassphrase"]); err != nil {
		return nil, err
	}
	dcrSetting, ok := confInfo["dcr"]
	if !ok {
		return nil, fmt.Errorf("Configure for btc wallet is required")
	}
	// setup dcr wallet
	dcrConf := asset.WalletConfig{
		Settings: dcrSetting,
		TipChange: func(e error) {

		},
	}
	dcrWallet, err := asset.Setup(42, &dcrConf, dex.NewLogger("DCR", dex.LevelTrace, os.Stdout), dex.Testnet)
	if err != nil {
		return nil, err
	}
	_, err = dcrWallet.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	if err = dcrWallet.Unlock(dcrSetting["walletpassphrase"]); err != nil {
		return nil, err
	}
	return &walletMatcher{
		btcWallet: btcWallet,
		dcrWallet: dcrWallet,
	}, nil
}

type initiateCmd struct {
	coin   string
	amount float64
}

func (c *initiateCmd) runCommand(ctx context.Context, wm *walletMatcher) error {
	var fromWallet, toWallet asset.Wallet
	switch c.coin {
	case "btc":
		fromWallet = wm.btcWallet
		toWallet = wm.dcrWallet
	case "dcr":
		fromWallet = wm.dcrWallet
		toWallet = wm.btcWallet
	default:
		return fmt.Errorf("Unsupported coin")
	}
	fmt.Sprintf("%v",toWallet)
	dexAsset := dex.Asset{
		ID:           0,
		Symbol:       c.coin,
		LotSize:      100000000,
		RateStep:     100000000,
		MaxFeeRate:   10,
		SwapSize:     0,
		SwapSizeBase: 0,
		SwapConf:     4,
	}
	qty := uint64(c.amount * 1e8)
	order := asset.Order{
		Value:        qty,
		MaxSwapCount: qty / dexAsset.LotSize,
		DEXConfig:    &dexAsset,
		Immediate:    true,
	}
	coins, redeemScript, err := fromWallet.FundOrder(&order)
	fmt.Println(coins, redeemScript, err)
	//err = fromWallet.ReturnCoins(coins)
	secret := encode.RandomBytes(32)
	secretHash := sha256.Sum256(secret)
	contract := asset.Contract{
		Address:    "",
		Value:      0,
		SecretHash: secretHash[:],
		LockTime:   uint64(time.Now().Add(time.Hour * 24).UTC().Unix()),
	}
	swap := asset.Swaps{
		Inputs:     coins,
		Contracts:  []*asset.Contract{&contract},
		FeeRate:    0,
		LockChange: false,
	}
	receipts,changeCoin, feePaid, err := fromWallet.Swap(&swap)
	// fromWallet.FindRedemption()
	fmt.Println(receipts,changeCoin, feePaid, err)
	return nil
}
