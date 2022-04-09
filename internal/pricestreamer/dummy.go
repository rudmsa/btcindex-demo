package pricestreamer

import (
	"context"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"

	"github.com/rudmsa/btcindex-demo/internal/exchange"
)

const (
	DummyStreamDefaultBuffer = 100
)

// Generates random TickerPrice according to specified low/high range with interval period
type dummyStream struct {
	tickerName exchange.Ticker
	low, high  decimal.Decimal
	interval   time.Duration

	timeTicker *time.Ticker
}

func NewDummyStream(l, h decimal.Decimal, in time.Duration) PriceStreamSubscriber {
	return &dummyStream{
		low:      l,
		high:     h,
		interval: in,
	}
}

func (ds *dummyStream) SubscribePriceStream(ctx context.Context, tick exchange.Ticker) (chan exchange.TickerPrice, chan error) {
	priceCh := make(chan exchange.TickerPrice, DummyStreamDefaultBuffer)
	errCh := make(chan error, 1)

	go func() {
		ds.tickerName = tick
		ds.timeTicker = time.NewTicker(ds.interval)

		tearDownFn := func() {
			close(priceCh)
			close(errCh)
			ds.timeTicker.Stop()
		}
		defer tearDownFn()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ds.timeTicker.C:
				select {
				case priceCh <- ds.generatePrice():
				default:
				}
			}
		}
	}()
	return priceCh, errCh
}

func (ds *dummyStream) generatePrice() exchange.TickerPrice {
	randomVal := decimal.NewFromFloat(rand.Float64())
	priceRange := ds.high.Sub(ds.low)
	price := decimal.Sum(randomVal.Mul(priceRange), ds.low)

	return exchange.TickerPrice{
		Ticker: ds.tickerName,
		Time:   time.Now(),
		Price:  price.String(),
	}
}