package facade

import (
	"testing"

	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingProfileRegistrar struct {
	profiles []httpx.HostProfile
}

func (r *recordingProfileRegistrar) UseProfile(profile httpx.HostProfile) {
	r.profiles = append(r.profiles, profile)
}

func TestYahooHostProfiles_ComposesQueryAndWebPolicies(t *testing.T) {
	config := httpx.DefaultConfig()
	config.QPS = 3.5
	config.Burst = 7
	config.MaxBodyBytes = 2 << 20
	config.UserAgent = "query-agent"

	recorder := &recordingProfileRegistrar{}
	registerYahooHostProfiles(recorder, config)
	require.Len(t, recorder.profiles, 4)

	profiles := make(map[string]httpx.HostProfile, len(recorder.profiles))
	for _, profile := range recorder.profiles {
		profiles[profile.Host] = profile
	}

	for _, host := range []string{
		"query1.finance.yahoo.com",
		"query2.finance.yahoo.com",
		"fc.yahoo.com",
	} {
		profile, ok := profiles[host]
		require.True(t, ok, "missing profile for %s", host)
		assert.Equal(t, 3.5, profile.QPS)
		assert.Equal(t, 7, profile.Burst)
		assert.EqualValues(t, 2<<20, profile.MaxBody)
		assert.Equal(t, "query-agent", profile.UserAgent)
	}
	assert.Equal(t, "bars_1d", profiles["query1.finance.yahoo.com"].EndpointFn("/v8/finance/chart/AAPL"))

	web := profiles["finance.yahoo.com"]
	assert.Equal(t, 0.7, web.QPS)
	assert.Equal(t, 1, web.Burst)
	assert.EqualValues(t, 8<<20, web.MaxBody)
	assert.Equal(t, "Mozilla/5.0 (Ampy yfinance-go scraper)", web.UserAgent)
	assert.Equal(t, "quote", web.EndpointFn("/quote/AAPL/financials"))
	assert.NotEmpty(t, web.Headers.Get("Accept"))
	assert.Equal(t, "en-US,en;q=0.9", web.Headers.Get("Accept-Language"))
}
