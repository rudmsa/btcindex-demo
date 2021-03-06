package main

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/rudmsa/btcindex-demo/internal/aggregator"
	"github.com/rudmsa/btcindex-demo/internal/core"
	"github.com/rudmsa/btcindex-demo/internal/indexer"
	"github.com/rudmsa/btcindex-demo/internal/indexer/algorithm"
	"github.com/rudmsa/btcindex-demo/internal/pricestreamer"
)

// nolint
func main() {
	exchangeNames := []string{"finhome-777", "coinhub-100", "cryptmaster-AAA", "futurebase-XXX", "cryptohub->>>"}

	testStreams := []pricestreamer.PriceStreamSubscriber{
		pricestreamer.NewDummyStream(decimal.NewFromFloat(35000), decimal.NewFromFloat(48000), 100*time.Millisecond),
		pricestreamer.NewDummyStream(decimal.NewFromFloat(45000), decimal.NewFromFloat(47000), 500*time.Millisecond),
		pricestreamer.NewDummyStream(decimal.NewFromFloat(45500), decimal.NewFromFloat(45600), 150*time.Millisecond),
		pricestreamer.NewDummyStream(decimal.NewFromFloat(45500), decimal.NewFromFloat(45600), 600*time.Millisecond),
	}

	btcAggregator := aggregator.NewAggregator(pricestreamer.BTCUSDTicker)
	for i := range testStreams {
		btcAggregator.RegisterPriceStreamer(exchangeNames[i], testStreams[i])
	}

	index := indexer.NewPriceIndexer(&algorithm.StreamingMean{}, 60*time.Second, btcAggregator.GetAggregatedOutput())

	var appl *core.Application = core.NewApplication()
	appl.Register(btcAggregator)
	appl.Register(index)

	ctx, cancelFn := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancelFn()

	go func() {
		appl.Run(ctx)
	}()

	fmt.Println("Timestamp, IndexPrice")
	for priceBar := range index.GetIndexOutput() {
		fmt.Printf("%d, %s\n", priceBar.Stamp, priceBar.Value.StringFixed(3))
	}
}
