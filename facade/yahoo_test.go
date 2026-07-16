package facade

import (
	"testing"

	"github.com/bizshuk/yfin/svc/yahoo"
	"github.com/stretchr/testify/require"
)

func TestProjectHolders(t *testing.T) {
	dto := &yahoo.HoldersDTO{
		MajorDirectHolders:   []yahoo.MajorDirectHolder{{Organization: "major"}},
		InstitutionOwnership: []yahoo.HolderRow{{Organization: "institution"}},
		FundOwnership:        []yahoo.HolderRow{{Organization: "fund"}},
	}

	major, err := projectHolders("major-holders", dto)
	require.NoError(t, err)
	require.Equal(t, dto.MajorDirectHolders, major)

	institution, err := projectHolders("institutional-holders", dto)
	require.NoError(t, err)
	require.Equal(t, dto.InstitutionOwnership, institution)

	fund, err := projectHolders("mutualfund-holders", dto)
	require.NoError(t, err)
	require.Equal(t, dto.FundOwnership, fund)

	_, err = projectHolders("unsupported", dto)
	require.EqualError(t, err, `unsupported holders command "unsupported"`)
}

func TestProjectInsider(t *testing.T) {
	dto := &yahoo.InsiderDTO{
		Transactions:     []yahoo.InsiderTransaction{{FilerName: "transaction"}},
		PurchaseActivity: yahoo.NetSharePurchaseActivity{Period: "6m"},
		Roster:           []yahoo.InsiderHolder{{Name: "roster"}},
	}

	transactions, err := projectInsider("insider-transactions", dto)
	require.NoError(t, err)
	require.Equal(t, dto.Transactions, transactions)

	purchases, err := projectInsider("insider-purchases", dto)
	require.NoError(t, err)
	require.Equal(t, yahoo.InsiderPurchaseSummaryTable(&dto.PurchaseActivity), purchases)

	roster, err := projectInsider("insider-roster", dto)
	require.NoError(t, err)
	require.Equal(t, dto.Roster, roster)

	_, err = projectInsider("unsupported", dto)
	require.EqualError(t, err, `unsupported insider command "unsupported"`)
}
