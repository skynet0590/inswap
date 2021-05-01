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
	"strings"
)

type walletMatcher struct {
	btcWallet asset.Wallet
	dcrWallet asset.Wallet
}

type command interface {
	runCommand(context.Context, *walletMatcher) error
}

var (
	flagset        = flag.NewFlagSet("", flag.ExitOnError)
	confFlag       = flagset.String("conf", "demo/config.json", "path to wallet connection config file")
	testnetFlag    = flagset.Bool("testnet", false, "use testnet network")
)

func init() {
	flagset.Usage = func() {
		fmt.Println("Usage: simpleswap [flags] cmd [cmd args]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  initiate <participant address> <amount>")
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
	switch
	args[0]{
	case "initiate":
		cmd = &initiateCmd{}
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
	data,err := ioutil.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}
	confInfo := make(map[string]map[string]string)
	err = json.Unmarshal(data, &confInfo)
	if err != nil {
		return nil, err
	}
	btcSetting,ok := confInfo["btc"]
	if !ok {
		return nil, fmt.Errorf("Configure for btc wallet is required")
	}
	// setup btc wallet
	btcConf := asset.WalletConfig{
		Settings:  btcSetting,
		TipChange: func(e error) {

		},
	}
	btcWallet, err := asset.Setup(0, &btcConf, dex.NewLogger("BTC", dex.LevelTrace, os.Stdout), dex.Testnet)
	if err != nil {
		return nil, err
	}
	_,err = btcWallet.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	if err = btcWallet.Unlock(btcSetting["walletpassphrase"]); err != nil {
		return nil, err
	}
	dcrSetting,ok := confInfo["dcr"]
	if !ok {
		return nil, fmt.Errorf("Configure for btc wallet is required")
	}
	// setup dcr wallet
	dcrConf := asset.WalletConfig{
		Settings:  dcrSetting,
		TipChange: func(e error) {

		},
	}
	dcrWallet, err := asset.Setup(42, &dcrConf, dex.NewLogger("DCR", dex.LevelTrace, os.Stdout), dex.Testnet)
	if err != nil {
		return nil, err
	}
	_,err = dcrWallet.Connect(context.Background())
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

}

func (c *initiateCmd) runCommand(ctx context.Context,wm *walletMatcher) error  {
	return nil
}
