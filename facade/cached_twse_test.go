package facade

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeTwseDispatcher struct {
	calls int
	value any
	err   error
}

func (f *fakeTwseDispatcher) Dispatch(context.Context, string, string, url.Values) (any, error) {
	f.calls++
	return f.value, f.err
}

func TestCachedTwseClientMissFetchesAndThenHitsCache(t *testing.T) {
	upstream := &fakeTwseDispatcher{value: map[string]any{
		"stat": "OK",
		"data": []any{[]any{"2330", "台積電"}},
	}}
	client, err := newCachedTwseClient(t.TempDir(), upstream)
	require.NoError(t, err)
	opts := url.Values{"selectType": {"ALL"}}

	first, err := client.Dispatch(context.Background(), "T86", "20260728", opts)
	require.NoError(t, err)
	require.Equal(t, upstream.value, first)
	require.Equal(t, 1, upstream.calls)

	upstream.err = errors.New("network must not be called")
	second, err := client.Dispatch(context.Background(), "T86", "20260728", url.Values{"selectType": {"ALL"}})
	require.NoError(t, err)
	require.Equal(t, 1, upstream.calls)
	require.Equal(t, "OK", second.(map[string]any)["stat"])
}

func TestCachedTwseClientCanonicalizesOptions(t *testing.T) {
	upstream := &fakeTwseDispatcher{value: map[string]any{"stat": "OK"}}
	client, err := newCachedTwseClient(t.TempDir(), upstream)
	require.NoError(t, err)

	first := url.Values{}
	first.Add("selectType", "ALL")
	first.Add("stockNo", "2330")
	second := url.Values{}
	second.Add("stockNo", "2330")
	second.Add("selectType", "ALL")

	_, err = client.Dispatch(context.Background(), "STOCK_DAY", "20260728", first)
	require.NoError(t, err)
	_, err = client.Dispatch(context.Background(), "STOCK_DAY", "20260728", second)
	require.NoError(t, err)
	require.Equal(t, 1, upstream.calls)
}
