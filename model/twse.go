// twse.go — TWSE (Taiwan Stock Exchange) API response types. Originally
// split across `svc/twse/types.go` (Response envelope) and 22 endpoint
// files (one `XxxResponse` struct + `XxxRow` typed row struct per endpoint).
// Consolidated here so any layer can reference these DTOs without pulling
// in svc/twse.
//
// Each endpoint `XxxResponse` embeds the common `Response` envelope and
// adds endpoint-specific fields (typically a `Date` string and optionally
// a `StockNo`). The `GetStat()` accessor on each one delegates to the
// embedded envelope's `Stat` field.
//
// Naming: TWSE convention `MI_INDEX` → Go `MI_INDEXResponse` (preserved
// verbatim). Row types use a more idiomatic PascalCase (`MIIndexRow` etc.).
// The embedded `model.Response` struct stays in this file.

package model

// Response is the common TWSE JSON envelope; concrete endpoints embed this
// and add their own extra fields (e.g. "date", "stockNo").
type Response struct {
	Stat   string     `json:"stat"`
	Title  string     `json:"title,omitempty"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
	Notes  []string   `json:"notes,omitempty"`
	Total  int        `json:"total,omitempty"`
	// Catch-all for endpoint-specific fields (date, stockNo, etc.) - decoded separately if needed.
	Extra map[string]any `json:"-"`
}

// GetStat exposes the embedded Stat field so callers can read it without
// importing reflection.
func (r *Response) GetStat() string { return r.Stat }

// MI_INDEXResponse embeds the common Response envelope and adds the
// `date` field that TWSE returns on this endpoint.
type MI_INDEXResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *MI_INDEXResponse) GetStat() string { return r.Response.Stat }

// MIIndexRow is a typed representation of one MI_INDEX data row.
// Fields: 指數, 收盤指數, 漲跌點數, 漲跌百分比.
type MIIndexRow struct {
	IndexName string  // 指數
	Close     float64 // 收盤指數
	Change    float64 // 漲跌點數
	ChangePct float64 // 漲跌百分比
}

// STOCK_DAYResponse embeds Response + per-stock fields.
type STOCK_DAYResponse struct {
	Response
	Date    string `json:"date"`
	StockNo string `json:"stockNo"`
}

// StockDayRow is one row of the STOCK_DAY table.
type StockDayRow struct {
	Date         string  // 日期
	Volume       int64   // 成交股數
	Amount       int64   // 成交金額
	Open         float64 // 開盤
	High         float64 // 最高
	Low          float64 // 最低
	Close        float64 // 收盤
	Change       float64 // 漲跌價差
	Transactions int64   // 成交筆數
}

// BWIBBU_dResponse embeds Response + date.
type BWIBBU_dResponse struct {
	Response
	Date string `json:"date"`
}

// BWIBBUdRow is one row of the BWIBBU_d (本益比、殖利率及股價淨值比) table.
type BWIBBUdRow struct {
	Code     string  // 證券代號
	Name     string  // 證券名稱
	PE       float64 // 本益比
	YieldPct float64 // 殖利率(%)
	PBR      float64 // 股價淨值比
}

// MI_INDEX_PLUSResponse — post-market fixed-price trading index.
type MI_INDEX_PLUSResponse struct {
	Response
	Date string `json:"date"`
}

// MIIndexPlusRow is one row of the MI_INDEX_PLUS table.
type MIIndexPlusRow struct {
	IndexName string  // 指數
	Close     float64 // 收盤指數
	Change    float64 // 漲跌點數
	ChangePct float64 // 漲跌百分比
}

// MI_INDEX_ODDResponse — odd-lot trading index.
type MI_INDEX_ODDResponse struct {
	Response
	Date string `json:"date"`
}

// MIIndexOddRow is one row of the MI_INDEX_ODD table.
type MIIndexOddRow struct {
	Code   string  // 證券代號
	Name   string  // 證券名稱
	Volume int64   // 成交股數
	Amount int64   // 成交金額
	Open   float64 // 開盤
	High   float64 // 最高
	Low    float64 // 最低
	Close  float64 // 收盤
}

// MI_5MINSResponse — every-5-second order/trade statistics.
type MI_5MINSResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *MI_5MINSResponse) GetStat() string { return r.Response.Stat }

// MI_5MINSRow is one row of the MI_5MINS table.
type MI_5MINSRow struct {
	Time           string // 時間
	CumBuyOrders   int64  // 累積委買筆數
	CumBuyLots     int64  // 累積委買張數
	CumSellOrders  int64  // 累積委賣筆數
	CumSellLots    int64  // 累積委賣張數
	CumTradeOrders int64  // 累計成交筆數
	CumTradeLots   int64  // 累計成交張數
}

// TWTB4UResponse — daily day-trade targets and statistics.
type TWTB4UResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *TWTB4UResponse) GetStat() string { return r.Response.Stat }

// TWTB4URow is one row of the TWTB4U table.
type TWTB4URow struct {
	Code        string // 證券代號
	Name        string // 證券名稱
	TradeShares int64  // 當日沖銷交易成交股數
	TradeAmount int64  // 當日沖銷交易成交金額
	BuyAmount   int64  // 當日沖銷交易買進成交金額
	SellAmount  int64  // 當日沖銷交易賣出成交金額
}

// MI_MARGNResponse — margin trading balance.
type MI_MARGNResponse struct {
	Response
	Date string `json:"date"`
}

// MI_MARGNRow is one row of the MI_MARGN table.
type MI_MARGNRow struct {
	Code          string // 股票代號
	Name          string // 股票名稱
	MarginBuy     int64  // 融資買進
	MarginSell    int64  // 融資賣出
	MarginRepay   int64  // 融資現償
	MarginBalance int64  // 融資餘額
	ShortBuy      int64  // 融券買進
	ShortSell     int64  // 融券賣出
	ShortRepay    int64  // 融券現償
	ShortBalance  int64  // 融券餘額
}

// T86Response — daily aggregated three-institution net buy/sell.
type T86Response struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *T86Response) GetStat() string { return r.Response.Stat }

// T86Row is one row of the T86 table.
type T86Row struct {
	Code        string // 證券代號
	Name        string // 證券名稱
	ForeignBuy  int64  // 外陸資買進股數
	ForeignSell int64  // 外陸資賣出股數
	ForeignNet  int64  // 外陸資買賣超股數
	TrustBuy    int64  // 投信買進股數
	TrustSell   int64  // 投信賣出股數
	TrustNet    int64  // 投信買賣超股數
	DealerBuy   int64  // 自營商買進股數
	DealerSell  int64  // 自營商賣出股數
	DealerNet   int64  // 自營商買賣超股數
	TotalNet    int64  // 三大法人買賣超股數
}

// MI_QFIISResponse — foreign-investor shareholding stats.
type MI_QFIISResponse struct {
	Response
	Date string `json:"date"`
}

// MI_QFIISRow is one row of the MI_QFIIS table.
type MI_QFIISRow struct {
	Code       string  // 證券代號
	Name       string  // 證券名稱
	SharesHeld int64   // 持有股數
	IssuePct   float64 // 佔發行股數%
}

// BFI82UResponse — three-institution aggregated buy/sell amount table.
type BFI82UResponse struct {
	Response
	Date string `json:"date"`
}

// BFI82URow is one row of the BFI82U table.
type BFI82URow struct {
	UnitName string  // 單位名稱
	Buy      float64 // 買進金額
	Sell     float64 // 賣出金額
	Net      float64 // 買賣差額
}

// TWT38UResponse — daily aggregated foreign-investor (含陸資) buy/sell volume.
type TWT38UResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *TWT38UResponse) GetStat() string { return r.Response.Stat }

// TWT38URow is one row of the TWT38U table.
type TWT38URow struct {
	UnitName string // 單位名稱
	Buy      int64  // 買進股數
	Sell     int64  // 賣出股數
	Net      int64  // 買賣差額股數
}

// TWT43UResponse — daily aggregated investment-trust (投信) buy/sell volume.
type TWT43UResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *TWT43UResponse) GetStat() string { return r.Response.Stat }

// TWT43URow is one row of the TWT43U table.
type TWT43URow struct {
	UnitName string // 單位名稱
	Buy      int64  // 買進股數
	Sell     int64  // 賣出股數
	Net      int64  // 買賣差額股數
}

// TWT44UResponse — daily aggregated dealer (自營商) buy/sell volume.
type TWT44UResponse struct {
	Response
	Date string `json:"date"`
}

// GetStat returns the embedded stat field.
func (r *TWT44UResponse) GetStat() string { return r.Response.Stat }

// TWT44URow is one row of the TWT44U table.
type TWT44URow struct {
	UnitName string // 單位名稱
	Buy      int64  // 買進股數
	Sell     int64  // 賣出股數
	Net      int64  // 買賣差額股數
}

// BlockBFIAUUResponse — block-trade daily details.
type BlockBFIAUUResponse struct {
	Response
	Date    string `json:"date"`
	StockNo string `json:"stockNo,omitempty"`
}

// BlockBFIAUURow is one row of the BlockBFIAUU table.
type BlockBFIAUURow struct {
	Seq           string  // 序號
	StockCode     string  // 證券代號
	StockName     string  // 證券名稱
	BuyBroker     string  // 買進證券商
	SellBroker    string  // 賣出證券商
	TradeVolume   int64   // 成交數量
	TradeAmount   float64 // 成交金額
	TradePrice    float64 // 成交價格
	TradeTime     string  // 成交時間
	BuyTradePrice float64 // 買進成交價
}

// BFIMUResponse — block-trade monthly summary.
type BFIMUResponse struct {
	Response
	Date string `json:"date"`
}

// BFIMUURow is one row of the BFIMUU monthly table.
type BFIMUURow struct {
	Period       string // 年月份
	Transactions int64  // 成交筆數
	Volume       int64  // 成交股數
	Amount       int64  // 成交金額
}

// BFIAUUYEARResponse — block-trade yearly summary.
type BFIAUUYEARResponse struct {
	Response
	Date string `json:"date"`
}

// BFIAUUYEARRow is one row of the BFIAUUYEAR table.
type BFIAUUYEARRow struct {
	Year         string // 年度
	Transactions int64  // 成交筆數
	Volume       int64  // 成交股數
	Amount       int64  // 成交金額
}

// FMTQIKResponse — TWSE index + trading volume for a given month.
type FMTQIKResponse struct {
	Response
	Date string `json:"date"`
}

// FMTQIKRow is one row of the FMTQIK table.
type FMTQIKRow struct {
	Date         string  // 日期
	Volume       int64   // 成交股數
	Amount       int64   // 成交金額
	Transactions int64   // 成交筆數
	Index        float64 // 發行量加權股價指數
}

// StockDayAvgResponse — per-stock monthly average price.
type StockDayAvgResponse struct {
	Response
	Date    string `json:"date"`
	StockNo string `json:"stockNo"`
}

// StockDayAvgRow is one row of the StockDayAvg monthly table.
type StockDayAvgRow struct {
	Year         string  // 年度
	Month        string  // 月份
	High         float64 // 最高
	Low          float64 // 最低
	WeightedAvg  float64 // 加權平均價
	Transactions int64   // 成交筆數
	Volume       int64   // 成交股數
	Amount       int64   // 成交金額
}

// FMSRFKResponse — per-stock monthly trading detail.
type FMSRFKResponse struct {
	Response
	StockNo string `json:"stockNo"`
	Date    string `json:"date"`
}

// FMSRFKRow is one row of the FMSRFK monthly table.
type FMSRFKRow struct {
	Year        string  // 年度
	Month       string  // 月份
	High        float64 // 最高
	Low         float64 // 最低
	WAvgPrice   float64 // 加權平均價
	TradeVolume int64   // 成交股數
	TradeValue  int64   // 成交金額
	TurnoverPct float64 // 週轉率%
}

// BFIAMUResponse — daily aggregated index trading volume/value.
type BFIAMUResponse struct {
	Response
	Date string `json:"date"`
}

// BFIAMURow is one row of the BFIAMU table.
type BFIAMURow struct {
	IndexName string  // 指數
	Close     float64 // 收盤指數
	Change    float64 // 漲跌
	ChangePct float64 // 百分比
}

// MI_WEEKResponse — weekly stock market value report.
type MI_WEEKResponse struct {
	Response
	Date string `json:"date"`
}

// MIWeekRow is one row of the MI_WEEK weekly table.
type MIWeekRow struct {
	StockCode    string // 股票代號
	StockName    string // 股票名稱
	SharesIssued int64  // 發行股數
	MarketCap    int64  // 市值
}