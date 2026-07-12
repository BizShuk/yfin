// profile_format.go — `printComprehensiveProfileSummary` formatter for the
// comprehensive-profile subcommand. See stats_format.go for the rationale on
// duplicating vs the scrape sub-package's copy.
package fundamentals

import (
	"fmt"

	"github.com/bizshuk/yfin/svc/scrape"
)

// printComprehensiveProfileSummary prints a summary of comprehensive profile
func printComprehensiveProfileSummary(dto *scrape.ComprehensiveProfileDTO) {
	fmt.Printf("COMPREHENSIVE PROFILE: symbol=%s\n", dto.Symbol)

	fmt.Printf("COMPANY INFORMATION:\n")
	if dto.CompanyName != "" {
		fmt.Printf("  Company Name: %s\n", dto.CompanyName)
	}
	if dto.ShortName != "" {
		fmt.Printf("  Short Name: %s\n", dto.ShortName)
	}
	if dto.Address1 != "" {
		fmt.Printf("  Address: %s\n", dto.Address1)
	}
	if dto.City != "" && dto.State != "" {
		fmt.Printf("  City, State: %s, %s\n", dto.City, dto.State)
	}
	if dto.Zip != "" {
		fmt.Printf("  ZIP: %s\n", dto.Zip)
	}
	if dto.Country != "" {
		fmt.Printf("  Country: %s\n", dto.Country)
	}
	if dto.Phone != "" {
		fmt.Printf("  Phone: %s\n", dto.Phone)
	}
	if dto.Website != "" {
		fmt.Printf("  Website: %s\n", dto.Website)
	}
	if dto.Industry != "" {
		fmt.Printf("  Industry: %s\n", dto.Industry)
	}
	if dto.Sector != "" {
		fmt.Printf("  Sector: %s\n", dto.Sector)
	}
	if dto.FullTimeEmployees != nil {
		fmt.Printf("  Full Time Employees: %d\n", *dto.FullTimeEmployees)
	}
	if dto.BusinessSummary != "" {
		summary := dto.BusinessSummary
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}
		fmt.Printf("  Business Summary: %s\n", summary)
	}

	if len(dto.Executives) > 0 {
		fmt.Printf("KEY EXECUTIVES:\n")
		for i, exec := range dto.Executives {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s", i+1, exec.Name)
			if exec.Title != "" {
				fmt.Printf(" - %s", exec.Title)
			}
			if exec.YearBorn != nil {
				fmt.Printf(" (Born: %d)", *exec.YearBorn)
			}
			if exec.TotalPay != nil {
				fmt.Printf(" - Total Pay: $%.2fM", float64(*exec.TotalPay)/1e6)
			}
			fmt.Printf("\n")
		}
	}

	fmt.Printf("ADDITIONAL INFORMATION:\n")
	if dto.MaxAge != nil {
		fmt.Printf("  Max Age: %d\n", *dto.MaxAge)
	}
	if dto.AuditRisk != nil {
		fmt.Printf("  Audit Risk: %d\n", *dto.AuditRisk)
	}
	if dto.BoardRisk != nil {
		fmt.Printf("  Board Risk: %d\n", *dto.BoardRisk)
	}
	if dto.CompensationRisk != nil {
		fmt.Printf("  Compensation Risk: %d\n", *dto.CompensationRisk)
	}
	if dto.ShareHolderRightsRisk != nil {
		fmt.Printf("  Share Holder Rights Risk: %d\n", *dto.ShareHolderRightsRisk)
	}
	if dto.OverallRisk != nil {
		fmt.Printf("  Overall Risk: %d\n", *dto.OverallRisk)
	}
}
