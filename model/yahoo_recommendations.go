// yahoo_recommendations.go — Yahoo analyst recommendation-trend DTO.
// Originally lived in svc/yahoo/recommendations.go; promoted to model/ so
// external consumers can depend on the shape without importing the
// Decode/Fetch behavior of svc/yahoo.

package model

// RecommendationTrendRow is one row from the recommendationTrend module.
type RecommendationTrendRow struct {
	Period     string `json:"period"`
	StrongBuy  int    `json:"strongBuy"`
	Buy        int    `json:"buy"`
	Hold       int    `json:"hold"`
	Sell       int    `json:"sell"`
	StrongSell int    `json:"strongSell"`
}