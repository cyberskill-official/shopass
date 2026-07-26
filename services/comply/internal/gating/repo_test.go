package gating

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPgRuleSource_LoadsCountryRulesIntoRegistry(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT country_code, gate_key, allowed").
		WillReturnRows(sqlmock.NewRows([]string{"country_code", "gate_key", "allowed", "value", "version"}).
			AddRow("VN", GateVoucherStacking, true, "allowed", 1).
			AddRow("MY", GateVoucherStacking, false, "deny", 1).
			AddRow("PH", GateVoucherStacking, false, "deny", 1))

	reg := NewRegistry(NewPgRuleSource(db))
	require.NoError(t, reg.Reload(context.Background()))

	require.True(t, reg.Allow("VN", GateVoucherStacking))
	require.False(t, reg.Allow("MY", GateVoucherStacking))
	require.False(t, reg.Allow("PH", GateVoucherStacking))
	require.False(t, reg.Allow("XX", GateVoucherStacking))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPgRuleSource_RejectsUnknownCountry(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT country_code, gate_key, allowed").
		WillReturnRows(sqlmock.NewRows([]string{"country_code", "gate_key", "allowed", "value", "version"}).
			AddRow("ZZ", GateVoucherStacking, true, "allowed", 1))

	_, err = NewPgRuleSource(db).Load(context.Background())
	require.ErrorContains(t, err, `unknown country "ZZ"`)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPgRuleSource_RejectsUnknownGate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT country_code, gate_key, allowed").
		WillReturnRows(sqlmock.NewRows([]string{"country_code", "gate_key", "allowed", "value", "version"}).
			AddRow("VN", "unknown_gate", true, "allowed", 1))

	_, err = NewPgRuleSource(db).Load(context.Background())
	require.ErrorContains(t, err, `unknown gate "unknown_gate"`)
	require.NoError(t, mock.ExpectationsWereMet())
}
