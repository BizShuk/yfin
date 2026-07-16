// Fetches/decodes a flattened `info` map by merging five quoteSummary modules (`assetProfile`, `summaryDetail`, `defaultKeyStatistics`, `financialData`, `quoteType`) and collapsing `{raw,...}` wrappers via `flattenValue`. Capacity: 1 var (`InfoModules`, 5 entries) + `DecodeInfo`, `FetchInfo`, internal `flattenValue`.
package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
)

// InfoModules are the quoteSummary modules merged into the flat .info map.
var InfoModules = []string{
	"assetProfile", "summaryDetail",
	"defaultKeyStatistics", "financialData", "quoteType",
}

func DecodeInfo(data []byte) (map[string]any, error) {
	var r struct {
		QuoteSummary struct {
			Result []map[string]json.RawMessage `json:"result"`
		} `json:"quoteSummary"`
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("info: empty result")
	}
	out := map[string]any{}
	for _, module := range InfoModules {
		modRaw, ok := r.QuoteSummary.Result[0][module]
		if !ok {
			continue
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(modRaw, &fields); err != nil {
			continue // module wasn't an object (e.g. null) — skip
		}
		for k, v := range fields {
			out[k] = flattenValue(v)
		}
	}
	return out, nil
}

// flattenValue collapses Yahoo {raw,...} objects to their raw value; other JSON kept as-is.
func flattenValue(v json.RawMessage) any {
	var obj struct {
		Raw json.RawMessage `json:"raw"`
	}
	if err := json.Unmarshal(v, &obj); err == nil && obj.Raw != nil {
		var raw any
		_ = json.Unmarshal(obj.Raw, &raw)
		return raw
	}
	var scalar any
	_ = json.Unmarshal(v, &scalar)
	return scalar
}

func (c *Client) FetchInfo(ctx context.Context, symbol string) (map[string]any, error) {
	raw, err := c.FetchQuoteSummary(ctx, symbol, InfoModules)
	if err != nil {
		return nil, err
	}
	return DecodeInfo(raw)
}
