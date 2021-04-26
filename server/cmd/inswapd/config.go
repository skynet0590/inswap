package main

import "github.com/decred/dcrd/dcrutil/v3"

var (
	defaultAppDataDir = dcrutil.AppDataDir("inswapd", false)
)

type (
	appConfig struct {
		DataDir           string
		DBName            string
		DBUser            string
		DBPass            string
		DBHost            string
		DBPort            uint16
	}
)
