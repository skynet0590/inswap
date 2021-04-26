package app

import (
	"fmt"
	"math"
	"strings"
)

// MarketInfo specifies a market that the Archiver must support.
type MarketInfo struct {
	Name                   string
	Base                   uint32
	Quote                  uint32
	LotSize                uint64
	EpochDuration          uint64 // msec
	MarketBuyBuffer        float64
	MaxUserCancelsPerEpoch uint32
	BookedLotLimit         uint32
}

func marketName(base, quote string) string {
	return base + "_" + quote
}

// MarketName creates the string representation of a DEX market (e.g. "dcr_btc")
// given the base and quote asset indexes defined in BIP-0044. See also
// BipIDSymbol.
func MarketName(base, quote uint32) (string, error) {
	baseSymbol := BipIDSymbol(base)
	if baseSymbol == "" {
		return "", fmt.Errorf("base asset %d not found", base)
	}
	baseSymbol = strings.ToLower(baseSymbol)

	quoteSymbol := BipIDSymbol(quote)
	if quoteSymbol == "" {
		return "", fmt.Errorf("quote asset %d not found", quote)
	}
	quoteSymbol = strings.ToLower(quoteSymbol)

	return marketName(baseSymbol, quoteSymbol), nil
}

// NewMarketInfo creates a new market configuration (MarketInfo) from the given
// base and quote asset indexes, order lot size, and epoch duration in
// milliseconds. See also MarketName.
func NewMarketInfo(base, quote uint32, lotSize, epochDuration uint64, marketBuyBuffer float64) (*MarketInfo, error) {
	name, err := MarketName(base, quote)
	if err != nil {
		return nil, err
	}
	return &MarketInfo{
		Name:                   name,
		Base:                   base,
		Quote:                  quote,
		LotSize:                lotSize,
		EpochDuration:          epochDuration,
		MarketBuyBuffer:        marketBuyBuffer,
		MaxUserCancelsPerEpoch: math.MaxUint32,
		BookedLotLimit:         math.MaxUint32,
	}, nil
}
