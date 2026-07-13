// quotes.go — `/v7/finance/quote` HTTP fetch + JSON decode + strict
// validation for the quote endpoint. Type definitions (QuoteResponse,
// QuoteResponseData, QuoteResult, RawQuote) live in `model/yahoo_raw.go`.
package yahoo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/bizshuk/yfin/model"
)

// Back-compat type aliases.
type (
	QuoteResponse     = model.QuoteResponse
	QuoteResponseData = model.QuoteResponseData
	QuoteResult       = model.QuoteResult
	Quote             = model.RawQuote
)

// DecodeQuoteResponse decodes a Yahoo Finance quote response with strict
// validation (bid <= ask, non-negative volume/size, no NaN/Inf prices).
func DecodeQuoteResponse(data []byte) (*model.QuoteResponse, error) {
	var response model.QuoteResponse
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %w", err)
	}
	if err := ValidateQuotes(&response); err != nil {
		return nil, fmt.Errorf("invalid quote response: %w", err)
	}
	return &response, nil
}

// DecodeQuoteResponseFromReader is the streaming variant.
func DecodeQuoteResponseFromReader(reader io.Reader) (*model.QuoteResponse, error) {
	var response model.QuoteResponse
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %w", err)
	}
	if err := ValidateQuotes(&response); err != nil {
		return nil, fmt.Errorf("invalid quote response: %w", err)
	}
	return &response, nil
}

// ValidateQuotes runs structural + value validation on a decoded quote
// response.
func ValidateQuotes(r *model.QuoteResponse) error {
	if r.QuoteResponse.Error != nil {
		return fmt.Errorf("yahoo finance error: %s", *r.QuoteResponse.Error)
	}
	if len(r.QuoteResponse.Result) == 0 {
		return errors.New("no quote results found")
	}
	for i, result := range r.QuoteResponse.Result {
		if err := validateQuoteResult(result); err != nil {
			return fmt.Errorf("result[%d]: %w", i, err)
		}
	}
	return nil
}

func validateQuoteResult(r model.QuoteResult) error {
	if r.Symbol == "" {
		return errors.New("missing symbol")
	}
	if r.Currency == "" {
		return errors.New("missing currency")
	}
	if r.Bid != nil && r.Ask != nil {
		if err := validatePrice(*r.Bid); err != nil {
			return fmt.Errorf("invalid bid price: %w", err)
		}
		if err := validatePrice(*r.Ask); err != nil {
			return fmt.Errorf("invalid ask price: %w", err)
		}
		if *r.Bid > *r.Ask {
			return fmt.Errorf("bid > ask: bid=%.4f, ask=%.4f", *r.Bid, *r.Ask)
		}
	}
	if r.BidSize != nil && *r.BidSize < 0 {
		return fmt.Errorf("negative bid size: %d", *r.BidSize)
	}
	if r.AskSize != nil && *r.AskSize < 0 {
		return fmt.Errorf("negative ask size: %d", *r.AskSize)
	}
	if r.RegularMarketPrice != nil {
		if err := validatePrice(*r.RegularMarketPrice); err != nil {
			return fmt.Errorf("invalid regular market price: %w", err)
		}
	}
	if r.RegularMarketOpen != nil {
		if err := validatePrice(*r.RegularMarketOpen); err != nil {
			return fmt.Errorf("invalid regular market open: %w", err)
		}
	}
	if r.RegularMarketDayHigh != nil {
		if err := validatePrice(*r.RegularMarketDayHigh); err != nil {
			return fmt.Errorf("invalid regular market high: %w", err)
		}
	}
	if r.RegularMarketDayLow != nil {
		if err := validatePrice(*r.RegularMarketDayLow); err != nil {
			return fmt.Errorf("invalid regular market low: %w", err)
		}
	}
	if r.RegularMarketVolume != nil && *r.RegularMarketVolume < 0 {
		return fmt.Errorf("negative regular market volume: %d", *r.RegularMarketVolume)
	}
	return nil
}

func validatePrice(price float64) error {
	if math.IsNaN(price) {
		return errors.New("NaN price")
	}
	if math.IsInf(price, 0) {
		return errors.New("infinite price")
	}
	if price < 0 {
		return fmt.Errorf("negative price: %.4f", price)
	}
	return nil
}