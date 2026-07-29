package facade

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/bizshuk/yfin/model"
	"github.com/stretchr/testify/require"
)

type fakeDailyBarsFetcher struct {
	calls int
	batch *BarBatch
	err   error
}

func (f *fakeDailyBarsFetcher) FetchDailyBars(
	context.Context,
	string,
	time.Time,
	time.Time,
	bool,
	string,
) (*BarBatch, error) {
	f.calls++
	return f.batch, f.err
}

func TestCachedClientMissFetchesAndThenHitsCache(t *testing.T) {
	upstream := &fakeDailyBarsFetcher{batch: &BarBatch{
		Symbol:         "AAPL",
		MIC:            "XNAS",
		FirstTradeDate: "1980-12-12",
		Bars: []model.Bar{{
			Date: "2026-07-28", Open: 210, High: 214, Low: 209, Close: 213, Volume: 42,
		}},
	}}
	client, err := newCachedClient(t.TempDir(), upstream)
	require.NoError(t, err)

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC)

	first, err := client.FetchDailyBars(context.Background(), "AAPL", start, end, true, "first")
	require.NoError(t, err)
	require.Equal(t, upstream.batch, first)
	require.Equal(t, "1980-12-12", first.FirstTradeDate)
	require.Equal(t, 1, upstream.calls)

	upstream.err = errors.New("network must not be called")
	second, err := client.FetchDailyBars(context.Background(), "AAPL", start, end, true, "second")
	require.NoError(t, err)
	require.Equal(t, upstream.batch, second)
	require.Equal(t, "1980-12-12", second.FirstTradeDate)
	require.Equal(t, 1, upstream.calls)
}

func TestDailyBarsCachePathIsVersioned(t *testing.T) {
	path, err := dailyBarsCachePath(
		"AAPL",
		time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC),
		true,
	)
	require.NoError(t, err)
	require.Equal(
		t,
		"yahoo/bars/v4/AAPL/20260701T000000Z_20260729T000000Z_true.json",
		filepath.ToSlash(path),
	)
}

func TestCachedClientRequestIdentityIncludesRangeAndAdjustment(t *testing.T) {
	upstream := &fakeDailyBarsFetcher{batch: &BarBatch{Symbol: "AAPL"}}
	client, err := newCachedClient(t.TempDir(), upstream)
	require.NoError(t, err)

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.FixedZone("offset", 8*60*60))
	end := time.Date(2026, 7, 29, 0, 0, 0, 0, time.FixedZone("offset", 8*60*60))
	_, err = client.FetchDailyBars(context.Background(), "AAPL", start, end, true, "")
	require.NoError(t, err)
	_, err = client.FetchDailyBars(context.Background(), "AAPL", start, end, false, "")
	require.NoError(t, err)
	_, err = client.FetchDailyBars(context.Background(), "AAPL", start.AddDate(0, 0, 1), end, true, "")
	require.NoError(t, err)

	require.Equal(t, 3, upstream.calls)
}

func TestCachedClientLoadDailyBarsMergesAndDeduplicates(t *testing.T) {
	upstream := &fakeDailyBarsFetcher{}
	client, err := newCachedClient(t.TempDir(), upstream)
	require.NoError(t, err)

	start1 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end1 := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	upstream.batch = &BarBatch{
		Symbol: "2330.TW",
		MIC:    "XTAI",
		Bars: []model.Bar{
			{Date: "2026-07-01", Close: 1000, CurrencyCode: "TWD"},
			{Date: "2026-07-15", Close: 1010, CurrencyCode: "TWD"},
		},
	}
	_, err = client.FetchDailyBars(context.Background(), "2330.TW", start1, end1, true, "")
	require.NoError(t, err)

	upstream.batch = &BarBatch{
		Symbol: "2330.TW",
		MIC:    "XTAI",
		Bars: []model.Bar{
			{Date: "2026-07-15", Close: 1015, CurrencyCode: "TWD"},
			{Date: "2026-07-28", Close: 1050, CurrencyCode: "TWD"},
		},
	}
	_, err = client.FetchDailyBars(context.Background(), "2330.TW", end1, end1.AddDate(0, 0, 13), true, "")
	require.NoError(t, err)

	got, err := client.LoadDailyBars("2330.TW")
	require.NoError(t, err)
	require.Equal(t, "2330.TW", got.Symbol)
	require.Equal(t, "XTAI", got.MIC)
	require.Equal(t, []string{"2026-07-01", "2026-07-15", "2026-07-28"}, []string{
		got.Bars[0].Date,
		got.Bars[1].Date,
		got.Bars[2].Date,
	})
	require.Equal(t, 1015.0, got.Bars[1].Close)
	require.Equal(t, 0, upstream.calls-2)
}

func TestCachedClientLoadDailyBarsMissing(t *testing.T) {
	client, err := newCachedClient(t.TempDir(), &fakeDailyBarsFetcher{})
	require.NoError(t, err)

	_, err = client.LoadDailyBars("MSFT")
	require.Error(t, err)
	require.True(t, errors.Is(err, fs.ErrNotExist))
}
