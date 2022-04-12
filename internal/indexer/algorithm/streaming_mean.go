package algorithm

import (
	"math"
	"sync"

	"github.com/shopspring/decimal"
)

type StreamingMean struct {
	mx      sync.Mutex
	value   decimal.Decimal
	samples int64
}

func (s *StreamingMean) AddValue(val decimal.Decimal) {
	s.mx.Lock()
	if s.samples == math.MaxInt64 {
		s.samples = 0
	}
	s.samples++

	delta := val.Sub(s.value)
	s.value = decimal.Sum(s.value, delta.Div(decimal.NewFromInt(s.samples)))

	s.mx.Unlock()
}

func (s *StreamingMean) Result() (decimal.Decimal, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if s.samples == 0 {
		return decimal.Decimal{}, ErrNoData
	}
	return s.value, nil
}
