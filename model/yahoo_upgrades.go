// yahoo_upgrades.go — Yahoo analyst upgrade/downgrade-history DTO.
// Originally lived in svc/yahoo/upgrades.go; promoted to model/ so external
// consumers can depend on the shape without importing the Decode/Fetch
// behavior of svc/yahoo.

package model

// UpgradeRow is one row from the upgradeDowngradeHistory module.
type UpgradeRow struct {
	EpochGradeDate int64  `json:"epochGradeDate"`
	Firm           string `json:"firm"`
	ToGrade        string `json:"toGrade"`
	FromGrade      string `json:"fromGrade"`
	Action         string `json:"action"`
}