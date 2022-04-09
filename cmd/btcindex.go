package main

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/rudmsa/btcindex-demo/internal/pricestreamer"
)

func main() {
	test := pricestreamer.NewDummyStream(decimal.NewFromFloat(35000.0), decimal.NewFromFloat(48000.0), 10*time.Millisecond)
	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	ch, _ := test.SubscribePriceStream(ctx, "Test")
	for data := range ch {
		fmt.Println(data)
	}
}
