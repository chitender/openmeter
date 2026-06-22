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

func TestCodeValidateFiat(t *testing.T) {
	require.NoError(t, currencyx.Code("USD").ValidateFiat())
	require.True(t, currencyx.Code("USD").IsFiat())

	err := currencyx.Code("CREDITS").ValidateFiat()
	require.Error(t, err)
	require.False(t, currencyx.Code("CREDITS").IsFiat())
}

func TestCodeValidateCustom(t *testing.T) {
	require.NoError(t, currencyx.Code("CREDITS").ValidateCustom())
	require.True(t, currencyx.Code("CREDITS").IsCustom())

	err := currencyx.Code("USD").ValidateCustom()
	require.ErrorContains(t, err, "custom currency code cannot conflict with fiat currency code")
	require.False(t, currencyx.Code("USD").IsCustom())
}

func TestCalculatorRequiresFiatCode(t *testing.T) {
	_, err := currencyx.Code("CREDITS").Calculator()
	require.Error(t, err)
}
