// — `ShouldSkip` decides whether a cached raw artifact is fresh enough to skip refetching; backed by `RefreshMap` mapping command → tier. Capacity: 5 tiers (daily/weekly/monthly/quarterly/annually), 32 commands mapped.

package cache

import (
	"path/filepath"
	"time"
)

// RefreshMap mirrors Python skills/scripts/config.py REFRESH_MAP.
var RefreshMap = map[string]string{
	"history": "daily", "recommendations": "daily", "recommendations-summary": "daily",
	"upgrades": "daily", "news": "daily", "metadata": "daily",
	"info": "monthly", "insider-transactions": "monthly", "insider-purchases": "monthly",
	"insider-roster": "monthly", "calendar": "monthly",
	"actions": "quarterly", "income": "quarterly", "balance": "quarterly", "cashflow": "quarterly",
	"major-holders": "quarterly", "institutional-holders": "quarterly", "mutualfund-holders": "quarterly",
	"earnings-dates": "quarterly", "earnings-history": "quarterly", "eps-trend": "quarterly",
	"eps-revisions": "quarterly", "earnings-estimates": "quarterly", "revenue-estimates": "quarterly",
	"growth-estimates": "quarterly", "price-targets": "quarterly", "sec-filings": "quarterly",
	"sustainability": "quarterly",
	"isin":           "annually", "options": "annually",
}

func quarter(m time.Month) int { return (int(m)-1)/3 + 1 }

// ShouldSkip returns true if a fresh enough cached file exists for (command, ticker)
// under rawDir, per the command's refresh tier. force=true bypasses cache.
func ShouldSkip(command, ticker string, force bool, rawDir string, now time.Time) bool {
	if force {
		return false
	}
	tier := RefreshMap[command]
	if tier == "" {
		tier = "daily"
	}
	matches, _ := filepath.Glob(filepath.Join(rawDir, command, ticker+".*.json"))
	var newest time.Time
	for _, f := range matches {
		base := filepath.Base(f)
		stem := base[:len(base)-len(".json")]
		if len(stem) <= len(ticker)+1 {
			continue
		}
		datePart := stem[len(ticker)+1:] // strip "TICKER."
		fd, err := time.Parse("2006-01-02", datePart)
		if err != nil {
			continue
		}
		if fd.After(newest) {
			newest = fd
		}
	}
	if newest.IsZero() {
		return false
	}
	switch tier {
	case "daily":
		return newest.Year() == now.Year() && newest.YearDay() == now.YearDay()
	case "weekly":
		return now.Sub(newest).Hours() < 7*24
	case "monthly":
		return newest.Year() == now.Year() && newest.Month() == now.Month()
	case "quarterly":
		return newest.Year() == now.Year() && quarter(newest.Month()) == quarter(now.Month())
	case "annually":
		return newest.Year() == now.Year()
	default:
		return false
	}
}
