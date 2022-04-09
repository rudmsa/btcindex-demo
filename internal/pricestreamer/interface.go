package pricestreamer

import (
	"context"

	"github.com/rudmsa/btcindex-demo/internal/exchange"
)

type PriceStreamSubscriber interface {
	SubscribePriceStream(ctx context.Context, tick exchange.Ticker) (chan exchange.TickerPrice, chan error)
}
