package aggregator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/rudmsa/btcindex-demo/internal/exchange"
	"github.com/rudmsa/btcindex-demo/internal/pricestreamer"
)

const (
	OutputDefaultBuffer = 100
)

var (
	ErrAlreadyStarted = errors.New("cannot be done when already started")
)

type TickerAggregator struct {
	ticker exchange.Ticker

	providers []*priceProvider
	muProvs   sync.Mutex

	output chan Quote

	isStarted bool
}

func NewAggregator(ticker exchange.Ticker) *TickerAggregator {
	return &TickerAggregator{
		ticker: ticker,
		output: make(chan Quote, OutputDefaultBuffer),
	}
}

func (aggr *TickerAggregator) GetAggregatedOutput() <-chan Quote {
	return aggr.output
}

func (aggr *TickerAggregator) RegisterPriceStreamer(source string, priceStreamer pricestreamer.PriceStreamSubscriber) error {
	if aggr.isStarted {
		return ErrAlreadyStarted
	}

	prov := priceProvider{
		source:   source,
		streamer: priceStreamer,
	}
	aggr.muProvs.Lock()
	aggr.providers = append(aggr.providers, &prov)
	aggr.muProvs.Unlock()

	return nil
}

func (aggr *TickerAggregator) Run(ctx context.Context) error {
	if aggr.isStarted {
		return ErrAlreadyStarted
	}

	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	wg := sync.WaitGroup{}
	wg.Add(len(aggr.providers))

	for _, prov := range aggr.providers {
		if err := aggr.startProvider(ctx, &wg, prov); err != nil {
			return fmt.Errorf("failed to start provider [%s]: %w", prov.source, err)
		}
	}
	aggr.isStarted = true

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
					prov.dataCh = nil
					break
				}
				quote, err := TickerPriceToQuote(prov.source, data)
				if err != nil {
					// #TODO: report this error
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
