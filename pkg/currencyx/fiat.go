package currencyx

import (
	"errors"

	"github.com/alpacahq/alpacadecimal"
	"github.com/invopop/gobl/currency"
)

// ValidateFiat validates that the code is a known ISO fiat currency.
func (c Code) ValidateFiat() error {
	if err := c.Validate(); err != nil {
		return err
	}

	return currency.Code(c).Validate()
}

func (c Code) IsFiat() bool {
	return c.ValidateFiat() == nil
}

// Calculator provides a fiat currency calculator object. This allows us not to
// resolve the fiat definition repeatedly, and callers can assume Def is valid.
func (c Code) Calculator() (Calculator, error) {
	if err := c.ValidateFiat(); err != nil {
		return Calculator{}, err
	}

	def := currency.Get(currency.Code(c))
	if def == nil {
		return Calculator{}, errors.New("currency definition is required")
	}

	return Calculator{
		Currency: c,
		Def:      def,
	}, nil
}

// TODO: Better name?!
type Calculator struct {
	Currency Code
	Def      *currency.Def
}

func (c Calculator) RoundToPrecision(amount alpacadecimal.Decimal) alpacadecimal.Decimal {
	// TODO: For now we are skipping the smallestDenomination, as that is a reference to the coins
	// in circulation, but should not be an issue for online payments.
	return amount.Round(int32(c.Def.Subunits))
}

func (c Calculator) Validate() error {
	var errs []error
	if err := c.Currency.ValidateFiat(); err != nil {
		errs = append(errs, err)
	}
	if c.Def == nil {
		errs = append(errs, errors.New("currency definition is required"))
	}
	return errors.Join(errs...)
}

func (c Calculator) IsRoundedToPrecision(amount alpacadecimal.Decimal) bool {
	return amount.Equal(c.RoundToPrecision(amount))
}
