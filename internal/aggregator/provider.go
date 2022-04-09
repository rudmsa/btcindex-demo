package aggregator

import (
	"github.com/rudmsa/btcindex-demo/internal/exchange"
	"github.com/rudmsa/btcindex-demo/internal/pricestreamer"
)

type priceProvider struct {
	source   string
	streamer pricestreamer.PriceStreamSubscriber
	dataCh   chan exchange.TickerPrice
	errCh    chan error
}
