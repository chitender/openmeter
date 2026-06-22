package currencyx

import "errors"

// ValidateCustom validates that the code can be used as a custom currency code
// and does not collide with a known ISO fiat currency code.
func (c Code) ValidateCustom() error {
	if err := c.Validate(); err != nil {
		return err
	}

	if c.IsFiat() {
		return errors.New("custom currency code cannot conflict with fiat currency code")
	}

	return nil
}

func (c Code) IsCustom() bool {
	return c.ValidateCustom() == nil
}
