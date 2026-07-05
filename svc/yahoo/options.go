// Fetches/decodes Yahoo option-chain data via `/v7/finance/options/{symbol}` (no crumb required). Capacity: ~5 structs (`OptionContract`, `OptionExpiry`, `OptionsDTO`, internal `optionsResult`) + `DecodeOptions`, `FetchOptions`.
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OptionContract struct {
	Strike            RawValue `json:"strike"`
	LastPrice         RawValue `json:"lastPrice"`
	Bid               RawValue `json:"bid"`
	Ask               RawValue `json:"ask"`
	Volume            RawInt   `json:"volume"`
	OpenInterest      RawInt   `json:"openInterest"`
	ImpliedVolatility RawValue `json:"impliedVolatility"`
}

type OptionExpiry struct {
	ExpirationDate int64            `json:"expirationDate"`
	Calls          []OptionContract `json:"calls"`
	Puts           []OptionContract `json:"puts"`
}

type OptionsDTO struct {
	ExpirationDates []int64        `json:"expirationDates"`
	Strikes         []float64      `json:"strikes"`
	Options         []OptionExpiry `json:"options"`
}

type optionsResult struct {
	OptionChain struct {
		Result []OptionsDTO `json:"result"`
	} `json:"optionChain"`
}

func DecodeOptions(data []byte) (*OptionsDTO, error) {
	var r optionsResult
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.OptionChain.Result) == 0 {
		return nil, fmt.Errorf("options: empty result")
	}
	d := r.OptionChain.Result[0]
	return &d, nil
}

func (c *Client) FetchOptions(ctx context.Context, symbol string) (*OptionsDTO, error) {
	u := c.baseURL + "/v7/finance/options/" + symbol
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return DecodeOptions(body)
}
