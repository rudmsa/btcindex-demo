package aggregator

import (
	"context"
	"fmt"
	"sync"

	"github.com/rudmsa/btcindex-demo/internal/model"
	"github.com/rudmsa/btcindex-demo/internal/pricestreamer"
)

const (
	OutputDefaultBuffer = 100
)

type TickerAggregator struct {
	ticker pricestreamer.Ticker

	providers []*priceProvider
	muProvs   sync.Mutex

	output chan model.Quote
}

func NewAggregator(ticker pricestreamer.Ticker) *TickerAggregator {
	return &TickerAggregator{
		ticker: ticker,
		output: make(chan model.Quote, OutputDefaultBuffer),
	}
}

func (aggr *TickerAggregator) GetAggregatedOutput() <-chan model.Quote {
	return aggr.output
}

func (aggr *TickerAggregator) RegisterPriceStreamer(source string, priceStreamer pricestreamer.PriceStreamSubscriber) {
	prov := priceProvider{
		source:   source,
		streamer: priceStreamer,
	}
	aggr.muProvs.Lock()
	aggr.providers = append(aggr.providers, &prov)
	aggr.muProvs.Unlock()
}

func (aggr *TickerAggregator) Run(ctx context.Context) error {
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	wg := sync.WaitGroup{}
	wg.Add(len(aggr.providers))

	for _, prov := range aggr.providers {
		if err := aggr.startProvider(ctx, &wg, prov); err != nil {
			return fmt.Errorf("failed to start provider [%s]: %w", prov.source, err)
		}
	}

	wg.Wait()
	return nil
}

func (aggr *TickerAggregator) startProvider(ctx context.Context, wg *sync.WaitGroup, prov *priceProvider) error {
	prov.dataCh, prov.errCh = prov.streamer.SubscribePriceStream(ctx, aggr.ticker)
	if prov.dataCh == nil || prov.errCh == nil {
		return fmt.Errorf("streamer [%s] returned nil channels", prov.source)
	}

	go func() {
		defer wg.Done()

	mainloop:
		for {
			select {
			case <-ctx.Done():
				prov.dataCh = nil
				prov.errCh = nil
				return

			case data, ok := <-prov.dataCh:
				if !ok {
					// provider error channel should provide reason
					prov.dataCh = nil
					break
				}
				quote, err := model.TickerPriceToQuote(prov.source, data)
				if err != nil {
					// #TODO: report this error
					break
				}

				select {
				case aggr.output <- quote:
				default:
					// #TODO: output channel is full - report it
				}

			case <-prov.errCh:
				// #TODO: we got error data - report it
				break mainloop
			}
		}
	}()
	return nil
}
