package indexer

import (
	"context"
	"time"

	"github.com/rudmsa/btcindex-demo/internal/indexer/algorithm"
	"github.com/rudmsa/btcindex-demo/internal/model"
)

type Indexer struct {
	interval time.Duration
	formula  algorithm.Formula
	input    <-chan model.Quote
	out      chan PriceBar
}

func NewPriceIndexer(formula algorithm.Formula, interval time.Duration, input <-chan model.Quote) *Indexer {
	return &Indexer{
		interval: interval,
		formula:  formula,
		input:    input,
		out:      make(chan PriceBar),
	}
}

func (ind *Indexer) GetIndexOutput() <-chan PriceBar {
	return ind.out
}

func (ind *Indexer) Run(ctx context.Context) error {
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	ind.mainLoop(ctx)
	close(ind.out)

	return nil
}

func (ind *Indexer) mainLoop(ctx context.Context) {
	ticker := time.NewTicker(nextTick(ind.interval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			ticker.Reset(nextTick(ind.interval))

			result, err := ind.formula.Result()
			if err != nil {
				// #FIXME: log error & continue
				break
			}

			select {
			case ind.out <- BuildPriceBar(result):
			default:
				// #FIXME report channel is full
			}

		case quote, ok := <-ind.input:
			if !ok {
				// #FIXME report input channel closing
				return
			}
			ind.formula.AddValue(quote.Price)
		}
	}
}

func nextTick(interval time.Duration) time.Duration {
	now := time.Now()
	expected := now.Round(interval).Add(interval)
	return expected.Sub(now)
}
