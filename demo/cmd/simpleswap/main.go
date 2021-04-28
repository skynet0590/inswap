package main

import (
	"context"
	"decred.org/dcrdex/client/asset"
	_ "decred.org/dcrdex/client/asset/btc"
	_ "decred.org/dcrdex/client/asset/dcr"
	"decred.org/dcrdex/dex"
	"flag"
	"fmt"
	"os"
)

type walletMatcher struct {
	btcWallet asset.Wallet
	dcrWallet asset.Wallet
}

type walletConf struct {
	RPCServer string `json:"rpcserver"`
	RPCUser string `json:"rpcuser"`
	RPCPass string `json:"rpcpass"`
	RPCCert string `json:"rpccert"`
}

type command interface {
	runCommand(context.Context, *walletMatcher) error
}

var (
	flagset        = flag.NewFlagSet("", flag.ExitOnError)
	confFlag       = flagset.String("conf", "demo/sampleconfig.json", "path to wallet connection config file")
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
	/*_ = flagset.Parse([]string{"conf", "testnet"})
	wm, err := initWallet(*confFlag)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
		return
	}
	var showUse bool*/
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
	fmt.Println(os.Args)
	err := flagset.Parse(os.Args[1:])
	_, err = initWallet(*confFlag)
	if err != nil {
		return err, false
	}

	args := flagset.Args()
	fmt.Println(args)
	if len(args) == 0 {
		return err, true
	}
	return nil, false
}

func initWallet(confFilePath string) (*walletMatcher, error) {
	fmt.Println(confFilePath)
	// setup btc wallet
	btcConf := asset.WalletConfig{
		Settings:  nil,
		TipChange: func(e error) {

		},
	}
	btcWallet, err := asset.Setup(0, &btcConf, dex.NewLogger("BTC", dex.LevelTrace, os.Stdout), dex.Testnet)
	if err != nil {
		return nil, err
	}
	// setup dcr wallet
	dcrConf := asset.WalletConfig{
		Settings:  nil,
		TipChange: func(e error) {

		},
	}
	dcrWallet, err := asset.Setup(42, &dcrConf, dex.NewLogger("BTC", dex.LevelTrace, os.Stdout), dex.Testnet)
	if err != nil {
		return nil, err
	}
	return &walletMatcher{
		btcWallet: btcWallet,
		dcrWallet: dcrWallet,
	}, nil
}
