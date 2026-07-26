package fraud

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

type execCall struct {
	sql  string
	args []any
}

type queryCall struct {
	sql  string
	args []any
}

type fakePGDB struct {
	rows    []pgx.Row
	queries []queryCall
	execs   []execCall
	execTag pgconn.CommandTag
	execErr error
}

func (f *fakePGDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	f.execs = append(f.execs, execCall{sql: sql, args: args})
	if f.execTag.String() == "" {
		f.execTag = pgconn.NewCommandTag("UPDATE 1")
	}
	return f.execTag, f.execErr
}

func (f *fakePGDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	f.queries = append(f.queries, queryCall{sql: sql, args: args})
	if len(f.rows) == 0 {
		return fakeRow{err: pgx.ErrNoRows}
	}
	row := f.rows[0]
	f.rows = f.rows[1:]
	return row
}

type fakeRow struct {
	values []any
	err    error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for i, d := range dest {
		switch out := d.(type) {
		case *int:
			*out = r.values[i].(int)
		default:
			return pgx.ErrNoRows
		}
	}
	return nil
}

func TestPGEventCounter_CountsRecentReferralRedeems(t *testing.T) {
	db := &fakePGDB{rows: []pgx.Row{fakeRow{values: []any{12}}}}
	counter := NewPGEventCounter(db)

	n, err := counter.CountRedeems(context.Background(), 42, 60)

	require.NoError(t, err)
	require.Equal(t, 12, n)
	require.Len(t, db.queries, 1)
	require.Contains(t, db.queries[0].sql, "referral_code")
	require.Contains(t, db.queries[0].sql, "referee.created_at")
	require.Equal(t, []any{int64(42), 60}, db.queries[0].args)
}

func TestPGClusterSizer_UsesRecursiveAccountLinkCTE(t *testing.T) {
	db := &fakePGDB{rows: []pgx.Row{fakeRow{values: []any{7}}}}
	sizer := NewPGClusterSizer(db)

	size, err := sizer.ClusterSize(context.Background(), 42)

	require.NoError(t, err)
	require.Equal(t, 7, size)
	require.Len(t, db.queries, 1)
	require.Contains(t, db.queries[0].sql, "WITH RECURSIVE reach")
	require.Contains(t, db.queries[0].sql, "account_link_edge")
	require.Equal(t, []any{int64(42)}, db.queries[0].args)
}

func TestPGSignalStore_UpsertsOpenSignalWithReasonsJSON(t *testing.T) {
	db := &fakePGDB{}
	store := NewPGSignalStore(db)

	err := store.UpsertOpen(context.Background(), 42, "combined", 75, []Reason{
		{Signal: "velocity", Detail: "12 referral redeems in 60 minutes exceeded threshold 10", Contribution: 40},
	})

	require.NoError(t, err)
	require.Len(t, db.execs, 1)
	require.Contains(t, db.execs[0].sql, "ON CONFLICT (subject_user_id, kind) DO UPDATE")
	require.Equal(t, int64(42), db.execs[0].args[0])
	require.Equal(t, "combined", db.execs[0].args[1])
	require.Equal(t, 75, db.execs[0].args[2])
	require.JSONEq(t, `[{"signal":"velocity","detail":"12 referral redeems in 60 minutes exceeded threshold 10","contribution":40}]`, db.execs[0].args[3].(string))
}

func TestPGAccountLinkStore_NormalizesReferralEdge(t *testing.T) {
	db := &fakePGDB{}
	store := NewPGAccountLinkStore(db)

	err := store.UpsertReferralEdge(context.Background(), 200, 100)

	require.NoError(t, err)
	require.Len(t, db.execs, 1)
	require.Contains(t, db.execs[0].sql, "account_link_edge")
	require.Equal(t, []any{int64(100), int64(200), "referral", 1.0}, db.execs[0].args)
}

func TestPGRewardHolder_LogsOnlyWhenPayoutHoldSchemaMissing(t *testing.T) {
	db := &fakePGDB{execErr: &pgconn.PgError{Code: "42P01"}}
	holder := NewPGRewardHolder(db, nil)

	err := holder.Hold(context.Background(), 42)

	require.NoError(t, err)
	require.Len(t, db.execs, 1)
	require.Contains(t, db.execs[0].sql, "payout_hold")
}
