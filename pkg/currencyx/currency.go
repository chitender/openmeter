package currencyx

import (
	"fmt"

	"github.com/invopop/gobl/currency"
)

// Code represents a durable currency code used by product, billing, and ledger
// paths. It can be either an ISO fiat currency code or a custom currency code.
type Code currency.Code

// Currency is the common contract for fiat and custom currencies.
type Currency interface {
	CurrencyCode() Code
	CurrencyType() CurrencyType
	Rounding() Rounding
	Calculator() (Calculator, error)
	Validate() error
}

// CurrencyType distinguishes custom currencies from ISO fiat currencies.
type CurrencyType string

const (
	CurrencyTypeCustom CurrencyType = "custom"
	CurrencyTypeFiat   CurrencyType = "fiat"
)

func (t CurrencyType) Validate() error {
	switch t {
	case CurrencyTypeCustom, CurrencyTypeFiat:
		return nil
	default:
		return fmt.Errorf("currency type: %s", t)
	}
}

func NewCurrency(code Code) (Currency, error) {
	if err := code.Validate(); err != nil {
		return nil, err
	}

	return code, nil
}

func (c Code) CurrencyCode() Code {
	return c
}

func (c Code) CurrencyType() CurrencyType {
	if currency.Get(currency.Code(c)) != nil {
		return CurrencyTypeFiat
	}

	return CurrencyTypeCustom
}

func (c Code) Rounding() Rounding {
	if def := currency.Get(currency.Code(c)); def != nil {
		return FiatRounding(int32(def.Subunits))
	}

	return DefaultCustomRounding()
}
