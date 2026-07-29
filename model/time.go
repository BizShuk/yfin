// time.go — UTC time and day-boundary conversion helpers for Yahoo epoch
// timestamps. Originally in svc/norm/time.go; promoted to model/ for reuse
// across normalize, emit, and external consumers.

package model

import (
	"fmt"
	"strings"
	"time"
)

// ToTradingDay derives the exchange-local calendar date represented by a
// Yahoo daily-bar timestamp and returns that local day's UTC boundaries.
func ToTradingDay(
	timestamp int64,
	exchangeTimezone string,
) (date string, start, end, eventTime time.Time, err error) {
	exchangeTimezone = strings.TrimSpace(exchangeTimezone)
	if exchangeTimezone == "" {
		return "", time.Time{}, time.Time{}, time.Time{},
			fmt.Errorf("exchange timezone is required")
	}
	location, err := time.LoadLocation(exchangeTimezone)
	if err != nil {
		return "", time.Time{}, time.Time{}, time.Time{},
			fmt.Errorf("load exchange timezone %q: %w", exchangeTimezone, err)
	}

	local := time.Unix(timestamp, 0).In(location)
	localStart := time.Date(
		local.Year(), local.Month(), local.Day(),
		0, 0, 0, 0, location,
	)
	localEnd := localStart.AddDate(0, 0, 1)
	return localStart.Format("2006-01-02"),
		localStart.UTC(),
		localEnd.UTC(),
		localEnd.UTC(),
		nil
}

// ToUTCDayBoundaries returns UTC boundaries for the timestamp's UTC day.
// Deprecated: daily market bars should use ToTradingDay with the exchange
// timezone so their calendar date remains tied to the trading venue.
func ToUTCDayBoundaries(timestamp int64) (start, end, eventTime time.Time) {
	_, start, end, eventTime, _ = ToTradingDay(timestamp, "UTC")
	return start, end, eventTime
}
