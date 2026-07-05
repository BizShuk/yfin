package yahoo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func int64Ptr(v int64) *int64       { return &v }
func float64Ptr(v float64) *float64 { return &v }

func TestDecodeInsider_ParsesTransactions(t *testing.T) {
	raw := []byte(`{"quoteSummary":{"result":[{
	  "insiderTransactions":{"transactions":[
	    {"filerName":"DOE JOHN","transactionText":"Sale at price 150.00",
	     "shares":{"raw":1000},"value":{"raw":150000},
	     "startDate":{"raw":1700000000}}]},
	  "netSharePurchaseActivity":{"period":"6m","buyInfoShares":{"raw":5000},
	     "sellInfoShares":{"raw":2000},"netInfoShares":{"raw":3000}},
	  "insiderHolders":{"holders":[
	    {"name":"DOE JANE","relation":"Director","positionDirect":{"raw":20000}}]}
	}],"error":null}}`)

	d, err := DecodeInsider(raw)
	require.NoError(t, err)
	require.Len(t, d.Transactions, 1)
	require.Equal(t, "DOE JOHN", d.Transactions[0].FilerName)
	require.NotNil(t, d.PurchaseActivity.NetInfoShares.Raw)
	require.Equal(t, int64(3000), *d.PurchaseActivity.NetInfoShares.Raw)
	require.Len(t, d.Roster, 1)
	require.Equal(t, "DOE JANE", d.Roster[0].Name)
}

func TestDecodeInsider_EmptyResult(t *testing.T) {
	_, err := DecodeInsider([]byte(`{"quoteSummary":{"result":[],"error":null}}`))
	require.Error(t, err)
}

func TestInsiderPurchaseSummaryTable(t *testing.T) {
	dto := &NetSharePurchaseActivity{
		Period:                   "6m",
		BuyInfoShares:            RawInt{Raw: int64Ptr(5000)},
		SellInfoShares:           RawInt{Raw: int64Ptr(2000)},
		NetInfoShares:            RawInt{Raw: int64Ptr(3000)},
		TotalInsiderShares:       RawInt{Raw: int64Ptr(100000)},
		NetPercentInsiderShares:  RawValue{Raw: float64Ptr(0.03)},
		BuyPercentInsiderShares:  RawValue{Raw: float64Ptr(0.05)},
		SellPercentInsiderShares: RawValue{Raw: float64Ptr(0.02)},
		BuyInfoCount:             RawInt{Raw: int64Ptr(10)},
		SellInfoCount:            RawInt{Raw: int64Ptr(5)},
		NetInfoCount:             RawInt{Raw: int64Ptr(5)},
	}
	tbl := InsiderPurchaseSummaryTable(dto)
	require.Equal(t, "Insider Purchases Last 6m", tbl.LabelColumn)
	require.Equal(t, []string{"Purchases", "Sales", "Net Shares Purchased (Sold)", "Total Insider Shares Held", "% Net Shares Purchased (Sold)", "% Buy Shares", "% Sell Shares"}, tbl.Labels)
	require.Equal(t, []any{int64(5000), int64(2000), int64(3000), int64(100000), 0.03, 0.05, 0.02}, tbl.Shares)
	require.Equal(t, []any{int64(10), int64(5), int64(5), nil, nil, nil, nil}, tbl.Trans)
}
