package currencyx

import (
	"errors"
	"strings"
)

const (
	MinCodeLength = 3
	MaxCodeLength = 24

	PostgresCodeSchemaType = "varchar(24)"
)

// Validate validates the structural form shared by fiat and custom currency
// codes. Semantic registry checks belong to product/API layers, not durable
// ledger facts.
func (c Code) Validate() error {
	value := string(c)
	if value == "" {
		return errors.New("currency code is required")
	}

	if strings.TrimSpace(value) != value {
		return errors.New("currency code cannot contain leading or trailing whitespace")
	}

	if len(value) < MinCodeLength || len(value) > MaxCodeLength {
		return errors.New("currency code must be between 3 and 24 characters")
	}

	if strings.Contains(value, "|") {
		return errors.New("currency code cannot contain route delimiter")
	}

	return nil
}
