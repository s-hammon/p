package p

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDateRange(t *testing.T) {
	dr := DateRange{
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	got := dr.formatMonthShortName()
	require.Equal(t, "Jan - Mar 2025", got)
	require.Equal(t, 59, dr.Days())
	require.Equal(t, 59*24, dr.Hours())

	dr = DateRange{
		time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	got = dr.formatMonthShortName()
	require.Equal(t, "Dec 2024 - Feb 2025", got)
	require.Equal(t, 62, dr.Days())
	require.Equal(t, 62*24, dr.Hours())
}
