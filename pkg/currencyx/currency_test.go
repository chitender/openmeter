package currencyx_test

import (
	"testing"

	"github.com/alpacahq/alpacadecimal"
	"github.com/stretchr/testify/require"

	"github.com/openmeterio/openmeter/pkg/currencyx"
)

func TestRoundToPrecision(t *testing.T) {
	cases := []struct {
		def      string
		amount   float64
		expected float64
	}{
		// Subunits = 2, smallestDenomination = 1
		{"USD", 1.23456789, 1.23},
		{"USD", 1.23556789, 1.24},

		// Subunits = 0, smallestDenomination = 1
		{"JPY", 1.23456789, 1.0},
		{"JPY", 1.9556789, 2.0},
	}

	for _, c := range cases {
		calculator, err := currencyx.Code(c.def).Calculator()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		amount := alpacadecimal.NewFromFloat(c.amount)
		result := calculator.RoundToPrecision(amount).InexactFloat64()

		require.Equal(t, c.expected, result)
	}
}

func TestCodeValidate(t *testing.T) {
	tests := []struct {
		name    string
		code    currencyx.Code
		wantErr string
	}{
		{
			name: "fiat code",
			code: currencyx.Code("USD"),
		},
		{
			name: "custom code",
			code: currencyx.Code("CREDITS"),
		},
		{
			name: "custom code with separator",
			code: currencyx.Code("AI_TOKENS"),
		},
		{
			name:    "empty",
			code:    currencyx.Code(""),
			wantErr: "currency code is required",
		},
		{
			name:    "too short",
			code:    currencyx.Code("XY"),
			wantErr: "currency code must be between 3 and 24 characters",
		},
		{
			name:    "too long",
			code:    currencyx.Code("THIS_CODE_IS_TOO_LONG_123"),
			wantErr: "currency code must be between 3 and 24 characters",
		},
		{
			name:    "routing delimiter",
			code:    currencyx.Code("BAD|CODE"),
			wantErr: "currency code cannot contain route delimiter",
		},
		{
			name:    "leading whitespace",
			code:    currencyx.Code(" USD"),
			wantErr: "currency code cannot contain leading or trailing whitespace",
		},
		{
			name:    "trailing whitespace",
			code:    currencyx.Code("USD "),
			wantErr: "currency code cannot contain leading or trailing whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.code.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestCodeCurrencyType(t *testing.T) {
	require.Equal(t, currencyx.CurrencyTypeFiat, currencyx.Code("USD").CurrencyType())
	require.Equal(t, currencyx.CurrencyTypeCustom, currencyx.Code("CREDITS").CurrencyType())
}

func TestCurrencyInterface(t *testing.T) {
	var _ currencyx.Currency = currencyx.Code("USD")

	tests := []struct {
		name     string
		currency currencyx.Currency
		code     currencyx.Code
		typ      currencyx.CurrencyType
	}{
		{
			name:     "generic fiat code",
			currency: currencyx.Code("USD"),
			code:     currencyx.Code("USD"),
			typ:      currencyx.CurrencyTypeFiat,
		},
		{
			name:     "generic custom code",
			currency: currencyx.Code("CREDITS"),
			code:     currencyx.Code("CREDITS"),
			typ:      currencyx.CurrencyTypeCustom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, tt.currency.Validate())
			require.Equal(t, tt.code, tt.currency.CurrencyCode())
			require.Equal(t, tt.typ, tt.currency.CurrencyType())
			require.NoError(t, tt.currency.Rounding().Validate())
		})
	}
}

func TestNewCurrency(t *testing.T) {
	fiat, err := currencyx.NewCurrency(currencyx.Code("USD"))
	require.NoError(t, err)
	require.Equal(t, currencyx.CurrencyTypeFiat, fiat.CurrencyType())

	custom, err := currencyx.NewCurrency(currencyx.Code("CREDITS"))
	require.NoError(t, err)
	require.Equal(t, currencyx.CurrencyTypeCustom, custom.CurrencyType())

	_, err = currencyx.NewCurrency(currencyx.Code("XY"))
	require.ErrorContains(t, err, "currency code must be between 3 and 24 characters")
}

func TestCalculatorSupportsCustomCurrency(t *testing.T) {
	calculator, err := currencyx.Code("CREDITS").Calculator()
	require.NoError(t, err)
	require.Equal(t, currencyx.CurrencyTypeCustom, calculator.CurrencyType())
	require.Nil(t, calculator.Def)
	require.Equal(t, currencyx.DefaultCustomRounding(), calculator.Rounding)

	amount := alpacadecimal.RequireFromString("2.5")
	require.Equal(t, "2", calculator.RoundToPrecision(amount).String())
	require.False(t, calculator.IsRoundedToPrecision(amount))
	require.True(t, calculator.IsRoundedToPrecision(alpacadecimal.NewFromInt(2)))
}

func TestCalculatorAppliesDefaultCustomBankersRounding(t *testing.T) {
	calculator, err := currencyx.Code("CREDITS").Calculator()
	require.NoError(t, err)

	tests := []struct {
		amount string
		want   string
	}{
		{amount: "2.49", want: "2"},
		{amount: "2.5", want: "2"},
		{amount: "2.51", want: "3"},
		{amount: "3.5", want: "4"},
		{amount: "-2.5", want: "-2"},
		{amount: "-3.5", want: "-4"},
	}

	for _, tt := range tests {
		t.Run(tt.amount, func(t *testing.T) {
			got := calculator.RoundToPrecision(alpacadecimal.RequireFromString(tt.amount))
			require.Equal(t, tt.want, got.String())
		})
	}
}

func TestRoundingDefaultsToBankersRounding(t *testing.T) {
	rounding := currencyx.Rounding{Precision: 2}

	got := rounding.Round(alpacadecimal.RequireFromString("1.225"))

	require.Equal(t, "1.22", got.String())
}

func TestCalculatorAppliesConfiguredCustomRounding(t *testing.T) {
	tests := []struct {
		name     string
		rounding currencyx.Rounding
		amount   string
		want     string
	}{
		{
			name: "custom precision defaults to bankers rounding",
			rounding: currencyx.Rounding{
				Precision: 2,
			},
			amount: "1.225",
			want:   "1.22",
		},
		{
			name: "custom precision rounds half to even upward when even digit is higher",
			rounding: currencyx.Rounding{
				Precision: 2,
			},
			amount: "1.235",
			want:   "1.24",
		},
		{
			name: "custom precision can use half up",
			rounding: currencyx.Rounding{
				Precision: 2,
				Mode:      currencyx.RoundingModeHalfUp,
			},
			amount: "1.225",
			want:   "1.23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculator, err := currencyx.NewCalculator(testCurrency{
				code:     currencyx.Code("CREDITS"),
				rounding: tt.rounding,
			})
			require.NoError(t, err)

			got := calculator.RoundToPrecision(alpacadecimal.RequireFromString(tt.amount))
			require.Equal(t, tt.want, got.String())
		})
	}
}

func TestCalculatorRejectsInvalidRounding(t *testing.T) {
	tests := []struct {
		name       string
		calculator currencyx.Calculator
		want       string
	}{
		{
			name: "negative precision",
			calculator: currencyx.Calculator{
				Currency: currencyx.Code("CREDITS"),
				Rounding: currencyx.Rounding{
					Precision: -1,
				},
			},
			want: "rounding precision must be non-negative",
		},
		{
			name: "invalid mode",
			calculator: currencyx.Calculator{
				Currency: currencyx.Code("CREDITS"),
				Rounding: currencyx.Rounding{
					Mode: currencyx.RoundingMode("ceil"),
				},
			},
			want: "rounding mode is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorContains(t, tt.calculator.Validate(), tt.want)
		})
	}

	_, err := currencyx.NewCalculator(testCurrency{
		code: currencyx.Code("CREDITS"),
		rounding: currencyx.Rounding{
			Mode: currencyx.RoundingMode("ceil"),
		},
	})
	require.ErrorContains(t, err, "rounding mode is invalid")
}

type testCurrency struct {
	code     currencyx.Code
	rounding currencyx.Rounding
}

func (c testCurrency) CurrencyCode() currencyx.Code {
	return c.code
}

func (c testCurrency) CurrencyType() currencyx.CurrencyType {
	return c.code.CurrencyType()
}

func (c testCurrency) Rounding() currencyx.Rounding {
	return c.rounding
}

func (c testCurrency) Calculator() (currencyx.Calculator, error) {
	return currencyx.NewCalculator(c)
}

func (c testCurrency) Validate() error {
	return c.code.Validate()
}
