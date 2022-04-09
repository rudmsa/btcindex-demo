package aggregator

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/rudmsa/btcindex-demo/internal/exchange"
)

type Quote struct {
	Ticker exchange.Ticker
	Source string
	Stamp  time.Time
	Price  decimal.Decimal
}

func TickerPriceToQuote(source string, tp exchange.TickerPrice) (Quote, error) {
	price, err := decimal.NewFromString(tp.Price)
	if err != nil {
		return Quote{}, fmt.Errorf("failed to convert price [%s]: %w", tp.Price, err)
	}
	return Quote{
		Ticker: tp.Ticker,
		Source: source,
		Stamp:  tp.Time,
		Price:  price,
	}, nil
}
