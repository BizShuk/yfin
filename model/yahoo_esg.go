// yahoo_esg.go — Yahoo ESG-score DTO.
// Originally lived in svc/yahoo/esg.go; promoted to model/ so external
// consumers can depend on the shape without importing the Decode/Fetch
// behavior of svc/yahoo.

package model

// ESGDTO is the decoded esgScores result for a single symbol.
type ESGDTO struct {
	TotalEsg           RawValue `json:"totalEsg"`
	EnvironmentScore   RawValue `json:"environmentScore"`
	SocialScore        RawValue `json:"socialScore"`
	GovernanceScore    RawValue `json:"governanceScore"`
	RatingYear         int      `json:"ratingYear"`
	HighestControversy RawValue `json:"highestControversy"`
}