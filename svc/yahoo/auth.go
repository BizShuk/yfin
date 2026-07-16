// `CrumbManager` handles Yahoo cookie + crumb authentication for quoteSummary. Capacity: 3 methods (`Crumb`, `Invalidate`, internal `bootstrapCookie`/`fetchCrumb`); one cached crumb, invalidated on 401 by `FetchQuoteSummary`.
package yahoo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/bizshuk/yfin/utils/httpx"
)

// CrumbManager handles Yahoo's cookie + crumb authentication.
type CrumbManager struct {
	httpClient *httpx.Client
	cookieURL  string // e.g. https://fc.yahoo.com
	apiBaseURL string // e.g. https://query2.finance.yahoo.com

	mu    sync.Mutex
	crumb string
}

func NewCrumbManager(httpClient *httpx.Client, cookieURL, apiBaseURL string) *CrumbManager {
	if cookieURL == "" {
		cookieURL = "https://fc.yahoo.com"
	}
	if apiBaseURL == "" {
		apiBaseURL = "https://query2.finance.yahoo.com"
	}
	return &CrumbManager{httpClient: httpClient, cookieURL: cookieURL, apiBaseURL: apiBaseURL}
}

// Crumb returns a cached crumb, fetching cookie+crumb on first use.
func (m *CrumbManager) Crumb(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.crumb != "" {
		return m.crumb, nil
	}
	if err := m.bootstrapCookie(ctx); err != nil {
		return "", err
	}
	crumb, err := m.fetchCrumb(ctx)
	if err != nil {
		return "", err
	}
	m.crumb = crumb
	return crumb, nil
}

// Invalidate clears the cached crumb (call on 401 to force re-fetch).
func (m *CrumbManager) Invalidate() {
	m.mu.Lock()
	m.crumb = ""
	m.mu.Unlock()
}

func (m *CrumbManager) bootstrapCookie(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", m.cookieURL+"/", nil)
	if err != nil {
		return err
	}
	resp, err := m.httpClient.Do(ctx, req)
	if err != nil {
		var statusErr *httpx.HTTPError
		if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusForbidden {
			return nil
		}
		return fmt.Errorf("cookie bootstrap failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // 403 acceptable; we only need Set-Cookie
	return nil
}

func (m *CrumbManager) fetchCrumb(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", m.apiBaseURL+"/v1/test/getcrumb", nil)
	if err != nil {
		return "", err
	}
	resp, err := m.httpClient.Do(ctx, req)
	if err != nil {
		return "", fmt.Errorf("getcrumb failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	crumb := strings.TrimSpace(string(body))
	if crumb == "" || strings.Contains(crumb, "<html") {
		return "", fmt.Errorf("empty or invalid crumb (consent flow may be required)")
	}
	return crumb, nil
}
