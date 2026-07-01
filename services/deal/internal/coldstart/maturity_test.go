package coldstart

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

type mockProduct struct {
	days int
}

func (m mockProduct) DaysOfHistory() int {
	return m.days
}

func fakeProduct(days int) mockProduct {
	return mockProduct{days: days}
}

func TestMaturity_Boundaries(t *testing.T) {
    cases := []struct {
        days int
        want State
    }{
        {13, StateNew}, {14, StateWarming}, {89, StateWarming}, {90, StateMature},
    }
    for _, c := range cases {
        require.Equal(t, c.want, Maturity(c.days), "days=%d", c.days)
    }
}

func TestIsFeatureReady_Gate90d(t *testing.T) {
    require.False(t, IsFeatureReady(fakeProduct(89)))
    require.True(t, IsFeatureReady(fakeProduct(90)))
    require.True(t, IsFeatureReady(fakeProduct(200)))
}

func TestPriorFor_ReturnsPriorWhenReady(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()
	
	repo := NewRepo(mock)
	
	mock.ExpectQuery("SELECT category_id, median_price, typical_discount_depth, sample_count FROM category_prior WHERE category_id = \\$1").
		WithArgs(int64(4221)).
		WillReturnRows(pgxmock.NewRows([]string{"category_id", "median_price", "typical_discount_depth", "sample_count"}).
			AddRow(int64(4221), int64(100000), 0.15, 50)) // sample_count >= 30
	
	cp, ok, err := repo.PriorFor(context.Background(), 4221)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, int64(100000), cp.MedianPrice)
}

func TestPriorFor_ReturnsFalseWhenSampleCountLow(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()
	
	repo := NewRepo(mock)
	
	mock.ExpectQuery("SELECT category_id, median_price, typical_discount_depth, sample_count FROM category_prior WHERE category_id = \\$1").
		WithArgs(int64(4221)).
		WillReturnRows(pgxmock.NewRows([]string{"category_id", "median_price", "typical_discount_depth", "sample_count"}).
			AddRow(int64(4221), int64(100000), 0.15, 29)) // sample_count < 30
	
	_, ok, err := repo.PriorFor(context.Background(), 4221)
	require.NoError(t, err)
	require.False(t, ok)
}
