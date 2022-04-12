package aggregator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/rudmsa/btcindex-demo/internal/model"
	"github.com/rudmsa/btcindex-demo/internal/pricestreamer"
)

const (
	OutputDefaultBuffer = 100

	StateStopped  = 0
	StateRunning  = 1
	StateStopping = 2
)

type priceProvider struct {
	source   string
	streamer pricestreamer.PriceStreamSubscriber
	dataCh   chan pricestreamer.TickerPrice
	errCh    chan error
}

type providerErrorMessage struct {
	prov *priceProvider
	err  error
}

type TickerAggregator struct {
	ticker pricestreamer.Ticker

	output chan model.Quote
	errCh  chan providerErrorMessage

	wg        sync.WaitGroup
	providers []*priceProvider
	muProvs   sync.Mutex

	state    int32
	ctx      context.Context
	cancelFn context.CancelFunc
}

func NewAggregator(ticker pricestreamer.Ticker) *TickerAggregator {
	return &TickerAggregator{
		ticker: ticker,
		output: make(chan model.Quote, OutputDefaultBuffer),
		errCh:  make(chan providerErrorMessage, 100),
	}
}

func (aggr *TickerAggregator) GetAggregatedOutput() <-chan model.Quote {
	return aggr.output
}

func (aggr *TickerAggregator) RegisterPriceStreamer(source string, priceStreamer pricestreamer.PriceStreamSubscriber) error {
	prov := priceProvider{
		source:   source,
		streamer: priceStreamer,
	}
	aggr.muProvs.Lock()
	aggr.providers = append(aggr.providers, &prov)
	aggr.muProvs.Unlock()

	if atomic.LoadInt32(&aggr.state) == StateRunning {
		return aggr.startProvider(aggr.ctx, &prov)
	}
	return nil
}

func (aggr *TickerAggregator) Stop() {
	if !atomic.CompareAndSwapInt32(&aggr.state, StateRunning, StateStopping) {
		return
	}
	aggr.cancelFn()
	aggr.wg.Wait()
	aggr.state = StateStopped
}

func (aggr *TickerAggregator) Run(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&aggr.state, StateStopped, StateRunning) {
		return errors.New("TickerAggregator is already started")
	}
	aggr.ctx, aggr.cancelFn = context.WithCancel(ctx)

	if err := aggr.startProviders(); err != nil {
		return fmt.Errorf("failed to start price providers: %w", err)
	}
	aggr.mainLoop()
	return nil
}

func (aggr *TickerAggregator) startProviders() error {
	aggr.muProvs.Lock()
	defer aggr.muProvs.Unlock()

	for _, prov := range aggr.providers {
		if err := aggr.startProvider(aggr.ctx, prov); err != nil {
			return fmt.Errorf("failed to start provider [%s]: %w", prov.source, err)
		}
	}
	return nil
}

func (aggr *TickerAggregator) mainLoop() {
	for {
		select {
		case <-aggr.ctx.Done():
			return

		case msg := <-aggr.errCh:
			// #TODO log error
			aggr.startProvider(aggr.ctx, msg.prov)
		}
	}
}

func (aggr *TickerAggregator) startProvider(ctx context.Context, prov *priceProvider) error {
	prov.dataCh, prov.errCh = prov.streamer.SubscribePriceStream(ctx, aggr.ticker)
	if prov.dataCh == nil || prov.errCh == nil {
		return fmt.Errorf("streamer [%s] returned nil channels", prov.source)
	}
	aggr.wg.Add(1)

	go func() {
		defer aggr.wg.Done()

		for {
			select {
			case <-ctx.Done():
				prov.dataCh = nil
				prov.errCh = nil
				return

			case data, ok := <-prov.dataCh:
				if !ok {
					// do one more iteration - error channel should provide reason
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

			case err, ok := <-prov.errCh:
				if !ok {
					prov.dataCh = nil
					prov.errCh = nil
					return
				}
				select {
				case aggr.errCh <- providerErrorMessage{prov, err}:
				default:
					// #TODO: error channel is full - report it
				}
				return
			}
		}
	}()
	return nil
}
