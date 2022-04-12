package algorithm

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrNoData = errors.New("no data")
)

type Formula interface {
	AddValue(decimal.Decimal)
	Result() (decimal.Decimal, error)
}
