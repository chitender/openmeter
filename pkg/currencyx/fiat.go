package currencyx

import (
	"errors"
	"fmt"

	"github.com/alpacahq/alpacadecimal"
	"github.com/invopop/gobl/currency"
)

// Calculator provides currency amount operations. Fiat currencies round to
// their ISO subunit precision; custom currencies use configurable rounding.
func (c Code) Calculator() (Calculator, error) {
	return NewCalculator(c)
}

func NewCalculator(c Currency) (Calculator, error) {
	if err := c.Validate(); err != nil {
		return Calculator{}, err
	}

	code := c.CurrencyCode()
	calculator := Calculator{
		Currency: code,
		Rounding: c.Rounding(),
	}

	if c.CurrencyType() == CurrencyTypeFiat {
		def := currency.Get(currency.Code(code))
		if def == nil {
			return Calculator{}, errors.New("currency definition is required")
		}

		calculator.Def = def
	}

	if err := calculator.Validate(); err != nil {
		return Calculator{}, err
	}

	return calculator, nil
}

type RoundingMode string

const (
	RoundingModeHalfUp   RoundingMode = "half_up"
	RoundingModeHalfEven RoundingMode = "half_even"
)

type Rounding struct {
	Precision int32
	Mode      RoundingMode
}

func FiatRounding(precision int32) Rounding {
	return Rounding{
		Precision: precision,
		Mode:      RoundingModeHalfUp,
	}
}

func DefaultCustomRounding() Rounding {
	return Rounding{
		Precision: 0,
		Mode:      RoundingModeHalfEven,
	}
}

func (r Rounding) Validate() error {
	var errs []error
	if r.Precision < 0 {
		errs = append(errs, errors.New("rounding precision must be non-negative"))
	}

	switch r.Mode {
	case "", RoundingModeHalfUp, RoundingModeHalfEven:
	default:
		errs = append(errs, fmt.Errorf("rounding mode is invalid: %s", r.Mode))
	}

	return errors.Join(errs...)
}

func (r Rounding) Round(amount alpacadecimal.Decimal) alpacadecimal.Decimal {
	switch r.Mode {
	case "", RoundingModeHalfEven:
		return amount.RoundBank(r.Precision)
	default:
		return amount.Round(r.Precision)
	}
}

func (r Rounding) RoundDown(amount alpacadecimal.Decimal) alpacadecimal.Decimal {
	return amount.RoundDown(r.Precision)
}

func (r Rounding) Unit() alpacadecimal.Decimal {
	return alpacadecimal.NewFromInt(1).Shift(-r.Precision)
}

// TODO: Better name?!
type Calculator struct {
	Currency Code
	Def      *currency.Def
	Rounding Rounding
}

func (c Calculator) CurrencyCode() Code {
	return c.Currency
}

func (c Calculator) CurrencyType() CurrencyType {
	return c.Currency.CurrencyType()
}

func (c Calculator) RoundToPrecision(amount alpacadecimal.Decimal) alpacadecimal.Decimal {
	return c.effectiveRounding().Round(amount)
}

func (c Calculator) Validate() error {
	var errs []error
	if err := c.Currency.Validate(); err != nil {
		errs = append(errs, err)
	}
	if c.Currency.CurrencyType() == CurrencyTypeFiat && c.Def == nil {
		errs = append(errs, errors.New("currency definition is required"))
	}
	if err := c.effectiveRounding().Validate(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (c Calculator) IsRoundedToPrecision(amount alpacadecimal.Decimal) bool {
	return amount.Equal(c.RoundToPrecision(amount))
}

func (c Calculator) effectiveRounding() Rounding {
	if c.Rounding.Mode != "" || c.Rounding.Precision != 0 {
		rounding := c.Rounding
		if rounding.Mode == "" {
			rounding.Mode = RoundingModeHalfEven
		}

		return rounding
	}

	if c.Def != nil {
		return FiatRounding(int32(c.Def.Subunits))
	}

	return DefaultCustomRounding()
}
