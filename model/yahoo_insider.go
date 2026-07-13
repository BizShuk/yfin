// yahoo_insider.go — Yahoo insider activity DTOs (transactions, purchase
// activity, roster) plus the yfinance-style summary-table shape.
// Originally lived in svc/yahoo/insider.go; promoted to model/ so external
// consumers can depend on the shapes without importing the Decode/Fetch
// behavior of svc/yahoo.

package model

// InsiderDTO is the decoded insiderTransactions + netSharePurchaseActivity
// + insiderHolders bundle for a single symbol.
type InsiderDTO struct {
	Transactions     []InsiderTransaction
	PurchaseActivity NetSharePurchaseActivity
	Roster           []InsiderHolder
}

// InsiderTransaction is one row from the insiderTransactions module.
type InsiderTransaction struct {
	FilerName       string `json:"filerName"`
	TransactionText string `json:"transactionText"`
	Shares          RawInt `json:"shares"`
	Value           RawInt `json:"value"`
	StartDate       RawInt `json:"startDate"`
	OwnershipType   string `json:"ownership"`
}

// NetSharePurchaseActivity summarises buys/sells over a single period.
type NetSharePurchaseActivity struct {
	Period                   string   `json:"period"`
	BuyInfoShares            RawInt   `json:"buyInfoShares"`
	SellInfoShares           RawInt   `json:"sellInfoShares"`
	NetInfoShares            RawInt   `json:"netInfoShares"`
	TotalInsiderShares       RawInt   `json:"totalInsiderShares"`
	NetPercentInsiderShares  RawValue `json:"netPercentInsiderShares"`
	BuyPercentInsiderShares  RawValue `json:"buyPercentInsiderShares"`
	SellPercentInsiderShares RawValue `json:"sellPercentInsiderShares"`
	BuyInfoCount             RawInt   `json:"buyInfoCount"`
	SellInfoCount            RawInt   `json:"sellInfoCount"`
	NetInfoCount             RawInt   `json:"netInfoCount"`
}

// InsiderHolder is a single insider-roster entry from the insiderHolders module.
type InsiderHolder struct {
	Name            string `json:"name"`
	Relation        string `json:"relation"`
	PositionDirect  RawInt `json:"positionDirect"`
	LatestTransDate RawInt `json:"latestTransDate"`
}

// InsiderPurchaseTable is one row in the yfinance-style label/value table.
type InsiderPurchaseTable struct {
	LabelColumn string
	Labels      []string
	Shares      []any
	Trans       []any
}