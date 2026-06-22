package currencyx

import (
	"github.com/invopop/gobl/currency"
)

// Code represents a durable currency code used by product, billing, and ledger
// paths. It can be either an ISO fiat currency code or a custom currency code.
type Code currency.Code
