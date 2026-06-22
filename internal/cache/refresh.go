package cache

import (
	"path/filepath"
	"time"
)

// RefreshMap mirrors Python yf/scripts/config.py REFRESH_MAP.
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
	"isin": "annually", "options": "annually",
}

func quarter(m time.Month) int { return (int(m) - 1) / 3 + 1 }

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
	var dates []time.Time
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
		dates = append(dates, fd)
	}
	if len(dates) == 0 {
		return false
	}
	for _, fd := range dates {
		var ok bool
		switch tier {
		case "daily":
			ok = fd.Year() == now.Year() && fd.YearDay() == now.YearDay()
		case "weekly":
			ok = now.Sub(fd).Hours() < 7*24
		case "monthly":
			ok = fd.Year() == now.Year() && fd.Month() == now.Month()
		case "quarterly":
			ok = fd.Year() == now.Year() && quarter(fd.Month()) == quarter(now.Month())
		case "annually":
			ok = fd.Year() == now.Year()
		}
		if !ok {
			return false
		}
	}
	return true
}
