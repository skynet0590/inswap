package main

import (
	"context"
	"decred.org/dcrdex/client/asset"
	_ "decred.org/dcrdex/client/asset/btc"
	_ "decred.org/dcrdex/client/asset/dcr"
	"decred.org/dcrdex/dex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type walletMatcher struct {
	btcWallet asset.Wallet
	dcrWallet asset.Wallet
}

type command interface {
	runCommand(ctx context.Context, party1WM, party2WM *walletMatcher) error
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
	case "swap":
		cmdArgs = 4
	case "refund":
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
	fromWm, toWm, err := initWallet(*confFlag)
	btcWL := fromWm.btcWallet
	balance,_ := btcWL.Balance()
	fmt.Printf("Balance(Available: %d, Immature: %d, Locked: %d) \n", balance.Available, balance.Immature, balance.Locked)
	balance,_ = toWm.btcWallet.Balance()
	fmt.Printf("Balance(Available: %d, Immature: %d, Locked: %d) \n", balance.Available, balance.Immature, balance.Locked)
	Locked := btcWL.Locked()
	fmt.Println("Locked: ", Locked)
	err = btcWL.Unlock("")
	fmt.Println(err)
	return nil, false
	if err != nil {
		return err, false
	}

	var cmd command
	switch args[0] {
	case "swap":
		fromAmount, err := strconv.ParseFloat(args[2], 64)
		toAmount, err := strconv.ParseFloat(args[4], 64)
		if err != nil {
			return fmt.Errorf("Amount must be number: ", err), false
		}
		cmd = &swapCmd{
			fromCoin:   args[1],
			fromAmount: fromAmount,
			toCoin: args[3],
			toAmount: toAmount,
		}
	case "refund":
		return fmt.Errorf("This func is not implemented yet"), false
	}
	err = cmd.runCommand(context.Background(), fromWm, toWm)
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

type configWM struct {
	Config     map[string]string `json:"config"`
	Passphrase string            `json:"passphrase"`
}

func newWalletMatcher(conf map[string]configWM) (*walletMatcher, error) {
	btcSetting, ok := conf["btc"]
	if !ok {
		return nil, fmt.Errorf("Configure for btc wallet is required")
	}
	// setup btc wallet
	btcConf := asset.WalletConfig{
		Settings: btcSetting.Config,
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
	if btcSetting.Passphrase != "" {
		if err = btcWallet.Unlock(btcSetting.Passphrase); err != nil {
			return nil, err
		}
	}
	dcrSetting, ok := conf["dcr"]
	if !ok {
		return nil, fmt.Errorf("Configure for dcr wallet is required")
	}
	// setup dcr wallet
	dcrConf := asset.WalletConfig{
		Settings: dcrSetting.Config,
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
	if dcrSetting.Passphrase != "" {
		if err = dcrWallet.Unlock(dcrSetting.Passphrase); err != nil {
			return nil, err
		}
	}

	return &walletMatcher{
		btcWallet: btcWallet,
		dcrWallet: dcrWallet,
	}, nil
}

func initWallet(confFilePath string) (fromWM, toWM *walletMatcher, err error) {
	data, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return nil, nil, err
	}
	confInfo := make(map[string]map[string]configWM)
	err = json.Unmarshal(data, &confInfo)
	if err != nil {
		return nil, nil, err
	}
	fromWM, err = newWalletMatcher(confInfo["party1"])
	if err != nil {
		return nil, nil, err
	}
	toWM, err = newWalletMatcher(confInfo["party2"])
	if err != nil {
		return nil, nil, err
	}
	return fromWM, toWM, err
}

