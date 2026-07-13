// yahoo_value.go — `RawValue` / `RawInt` nullable wrappers around Yahoo's
// `{raw, fmt, longFmt}` value objects. Originally lived in svc/yahoo/rawvalue.go;
// promoted to model/ so the Yahoo Finance DTOs (esg / holders / insider / etc.)
// can reference the wrapper shapes without importing svc/yahoo for any behavior.

package model

// RawValue models Yahoo's {raw, fmt, longFmt} value object (float-flavoured).
type RawValue struct {
	Raw     *float64 `json:"raw,omitempty"`
	Fmt     string   `json:"fmt,omitempty"`
	LongFmt string   `json:"longFmt,omitempty"`
}

// RawInt is the integer-flavoured variant (timestamps, share counts).
type RawInt struct {
	Raw     *int64 `json:"raw,omitempty"`
	Fmt     string `json:"fmt,omitempty"`
	LongFmt string `json:"longFmt,omitempty"`
}