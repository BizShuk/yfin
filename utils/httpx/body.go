// body.go — Gzip auto-decode + per-response body size cap shared by `Caller.Get`.
// Capacity: 1 sentinel, 1 helper.
package httpx

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrBodyTooLarge is returned by readBody (and therefore Caller.Get) when a
// response body, after gzip decoding, exceeds Config.MaxBodyBytes.
var ErrBodyTooLarge = errors.New("httpx: response body exceeds MaxBodyBytes")

// readBody consumes resp.Body, transparently decoding gzip when the
// Content-Encoding header advertises it, and applying a hard byte cap when
// maxBytes > 0. The caller must not use resp.Body after this returns.
func readBody(resp *http.Response, maxBytes int64) ([]byte, error) {
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("httpx: gzip reader: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	// Read one byte past the limit so we can distinguish "exactly at limit"
	// from "exceeded limit".
	if maxBytes > 0 {
		reader = io.LimitReader(reader, maxBytes+1)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("httpx: read body: %w", err)
	}

	if maxBytes > 0 && int64(len(body)) > maxBytes {
		return nil, ErrBodyTooLarge
	}
	return body, nil
}
