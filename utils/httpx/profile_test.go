package httpx

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeProfileHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		host string
		want string
	}{
		{name: "lowercase", host: "QUERY1.FINANCE.YAHOO.COM", want: "query1.finance.yahoo.com"},
		{name: "trailing dot", host: "finance.yahoo.com.", want: "finance.yahoo.com"},
		{name: "port", host: "Example.COM.:8443", want: "example.com"},
		{name: "empty", host: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, normalizeProfileHost(tt.host))
		})
	}
}

func TestUseProfileRegistersDefensiveCopy(t *testing.T) {
	t.Parallel()

	client := NewClient(nil)
	headers := http.Header{"X-Profile": []string{"original"}}
	client.UseProfile(HostProfile{
		Host:       "Finance.Yahoo.COM.:443",
		EndpointFn: CommonEndpointFn,
		QPS:        0.7,
		Burst:      1,
		MaxBody:    8 << 20,
		UserAgent:  "profile-agent",
		Headers:    headers,
	})

	headers.Set("X-Profile", "mutated-source")
	profile, ok := client.profileForHost("FINANCE.YAHOO.COM")
	require.True(t, ok)
	assert.Equal(t, "finance.yahoo.com", profile.Host)
	assert.Equal(t, "original", profile.Headers.Get("X-Profile"))

	profile.Headers.Set("X-Profile", "mutated-result")
	again, ok := client.profileForHost("finance.yahoo.com:443")
	require.True(t, ok)
	assert.Equal(t, "original", again.Headers.Get("X-Profile"))
}

func TestUseProfileIgnoresInvalidHost(t *testing.T) {
	t.Parallel()

	client := NewClient(nil)
	client.UseProfile(HostProfile{Host: ""})

	_, ok := client.profileForHost("")
	assert.False(t, ok)
}

func TestUseProfileReplacesProfileAndUnknownHostFallsBack(t *testing.T) {
	t.Parallel()

	client := NewClient(nil)
	client.UseProfile(HostProfile{Host: "example.test", QPS: 1, Burst: 1})
	client.UseProfile(HostProfile{Host: "EXAMPLE.TEST.", QPS: 2.5, Burst: 4})

	profile, ok := client.profileForHost("example.test:443")
	require.True(t, ok)
	assert.Equal(t, 2.5, profile.QPS)
	assert.Equal(t, 4, profile.Burst)

	_, ok = client.profileForHost("unknown.test")
	assert.False(t, ok)
}
