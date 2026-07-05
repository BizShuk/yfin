package norm

import "time"

type NormalizedInsiderTxn struct {
	FilerName string     `json:"filer_name"`
	Text      string     `json:"text"`
	Shares    *int64     `json:"shares,omitempty"`
	Value     *int64     `json:"value,omitempty"`
	Date      *time.Time `json:"date,omitempty"`
}

type NormalizedInsider struct {
	Security     Security               `json:"security"`
	Transactions []NormalizedInsiderTxn `json:"transactions"`
	NetBuyShares *int64                 `json:"net_buy_shares,omitempty"`
	AsOf         time.Time              `json:"as_of"`
	Meta         Meta                   `json:"meta"`
}
