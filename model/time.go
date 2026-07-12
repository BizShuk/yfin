// time.go — UTC time and day-boundary conversion helpers for Yahoo epoch
// timestamps. Originally in svc/norm/time.go; promoted to model/ for reuse
// across normalize, emit, and external consumers.

package model

import "time"

// ToUTCDayBoundaries converts a Unix timestamp to UTC day boundaries for daily bars.
// For daily bars: start = 00:00:00Z of the day, end = next day 00:00:00Z,
// event_time = end. The timestamp represents the end of the trading day, so
// we map it to the trading day by subtracting two days (Yahoo's epoch is
// offset 2 days from the actual trading session in UTC).
func ToUTCDayBoundaries(timestamp int64) (start, end, eventTime time.Time) {
	t := time.Unix(timestamp, 0).UTC()
	tradingDay := t.AddDate(0, 0, -2)

	start = time.Date(tradingDay.Year(), tradingDay.Month(), tradingDay.Day(), 0, 0, 0, 0, time.UTC)
	end = start.Add(24 * time.Hour)
	eventTime = end

	return start, end, eventTime
}
