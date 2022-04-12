package pricestreamer

import (
	"context"
	"time"
)

type Ticker string

const (
	BTCUSDTicker Ticker = "BTC_USD"
)

type TickerPrice struct {
	Ticker Ticker
	Time   time.Time
	Price  string // decimal value. example: "0", "10", "12.2", "13.2345122"
}

type PriceStreamSubscriber interface {
	SubscribePriceStream(ctx context.Context, tick Ticker) (chan TickerPrice, chan error)
}
