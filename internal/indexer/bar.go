package indexer

import (
	"time"

	"github.com/shopspring/decimal"
)

type PriceBar struct {
	Stamp int64
	Value decimal.Decimal
}

func BuildPriceBar(val decimal.Decimal) PriceBar {
	return PriceBar{
		Stamp: time.Now().Unix(),
		Value: val,
	}
}
