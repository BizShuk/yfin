// yahoo_secfilings.go — Yahoo SEC-filing metadata DTO.
// Originally lived in svc/yahoo/secfilings.go; promoted to model/ so
// external consumers can depend on the shape without importing the
// Decode/Fetch behavior of svc/yahoo.

package model

// SecFiling is one row from the secFilings module.
type SecFiling struct {
	Date      string `json:"date"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	EdgarURL  string `json:"edgarUrl"`
	EpochDate int64  `json:"epochDate"`
}