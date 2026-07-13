// yahoo_holders.go — Yahoo major-holders + institution/fund ownership DTOs.
// Originally lived in svc/yahoo/holders.go; promoted to model/ so external
// consumers can depend on the shapes without importing the Decode/Fetch
// behavior of svc/yahoo.

package model

// HoldersDTO is the decoded majorHoldersBreakdown + ownership result for a symbol.
type HoldersDTO struct {
	MajorBreakdown       MajorHoldersBreakdown
	MajorDirectHolders   []MajorDirectHolder
	InstitutionOwnership []HolderRow
	FundOwnership        []HolderRow
}

// MajorHoldersBreakdown is the aggregate insider/institution holding percentages.
type MajorHoldersBreakdown struct {
	InsidersPercentHeld      RawValue `json:"insidersPercentHeld"`
	InstitutionsPercentHeld  RawValue `json:"institutionsPercentHeld"`
	InstitutionsFloatPctHeld RawValue `json:"institutionsFloatPercentHeld"`
	InstitutionsCount        RawInt   `json:"institutionsCount"`
}

// MajorDirectHolder is a single row from the majorDirectHolders module.
type MajorDirectHolder struct {
	Organization       string `json:"organization"`
	PositionDirect     RawInt `json:"positionDirect"`
	PositionDirectDate RawInt `json:"positionDirectDate"`
	ValueDirect        RawInt `json:"valueDirect"`
}

// HolderRow is one institution/fund ownership record.
type HolderRow struct {
	Organization string   `json:"organization"`
	PctHeld      RawValue `json:"pctHeld"`
	Position     RawInt   `json:"position"`
	Value        RawInt   `json:"value"`
}