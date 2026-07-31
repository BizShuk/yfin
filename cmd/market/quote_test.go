package market

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateQuoteFlagsRequiresCompleteExportPair(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  quoteConfig
		wantErr string
	}{
		{
			name:    "format without directory",
			config:  quoteConfig{Tickers: "AAPL", Out: "json"},
			wantErr: "--out-dir is required when --out is set",
		},
		{
			name:    "directory without format",
			config:  quoteConfig{Tickers: "AAPL", OutDir: "./quotes"},
			wantErr: "--out is required when --out-dir is set",
		},
		{
			name:   "complete pair",
			config: quoteConfig{Tickers: "AAPL", Out: "json", OutDir: "./quotes"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateQuoteFlags(&test.config)
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, test.wantErr)
		})
	}
}
