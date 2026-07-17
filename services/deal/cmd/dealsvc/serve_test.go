package main

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestChartRepoChecksUserProductLink(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery(`SELECT EXISTS\( SELECT 1 FROM user_tracked_product WHERE user_id = \$1 AND product_id = \$2 \)`).
		WithArgs(int64(123), int64(77)).
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	repo := &chartRepo{pool: mock}
	allowed, err := repo.UserCanViewProduct(context.Background(), 123, 77)
	require.NoError(t, err)
	require.True(t, allowed)
	require.NoError(t, mock.ExpectationsWereMet())
}
