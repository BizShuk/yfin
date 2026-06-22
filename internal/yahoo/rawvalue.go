package yahoo

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
