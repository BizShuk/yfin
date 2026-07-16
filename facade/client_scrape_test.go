package facade

import (
	"testing"

	"github.com/bizshuk/yfin/model"
	"github.com/stretchr/testify/require"
)

func TestProjectAnalysisDimension(t *testing.T) {
	dto := &model.ComprehensiveAnalysisDTO{}
	dto.EarningsHistory.Currency = "history"
	dto.EPSTrend.Currency = "trend"
	dto.EPSRevisions.Currency = "revisions"
	dto.EarningsEstimate.Currency = "earnings"
	dto.RevenueEstimate.Currency = "revenue"
	growth := "growth"
	dto.GrowthEstimate.CurrentQtr = &growth

	tests := []struct {
		command string
		want    any
	}{
		{command: "earnings-history", want: dto.EarningsHistory},
		{command: "eps-trend", want: dto.EPSTrend},
		{command: "eps-revisions", want: dto.EPSRevisions},
		{command: "earnings-estimates", want: dto.EarningsEstimate},
		{command: "revenue-estimates", want: dto.RevenueEstimate},
		{command: "growth-estimates", want: dto.GrowthEstimate},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got, err := projectAnalysisDimension(tt.command, dto)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	_, err := projectAnalysisDimension("unsupported", dto)
	require.EqualError(t, err, `unsupported analysis command "unsupported"`)
}
