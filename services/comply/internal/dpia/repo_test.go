package dpia

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*Service, *pgxpool.Pool) {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), `
		DROP TABLE IF EXISTS tia CASCADE;
		DROP TABLE IF EXISTS dpia CASCADE;
		DROP TABLE IF EXISTS processing_activity CASCADE;

		CREATE TABLE processing_activity (
			id               BIGSERIAL   PRIMARY KEY,
			name             TEXT        NOT NULL,
			purpose_key      TEXT        NOT NULL,
			data_categories  TEXT[]      NOT NULL,
			started_at       TIMESTAMPTZ NOT NULL,
			cross_border     BOOLEAN     NOT NULL DEFAULT false,
			recipient_country TEXT,
			created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
			CHECK (NOT cross_border OR recipient_country IS NOT NULL)
		);

		CREATE TABLE dpia (
			id               BIGSERIAL   PRIMARY KEY,
			activity_id      BIGINT      NOT NULL REFERENCES processing_activity(id),
			version          INTEGER     NOT NULL,
			risk_level       TEXT        NOT NULL CHECK (risk_level IN ('low','medium','high')),
			mitigation_vi    TEXT,
			status           TEXT        NOT NULL DEFAULT 'draft',
			filed_at         TIMESTAMPTZ,
			last_reviewed_at TIMESTAMPTZ,
			created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
			UNIQUE (activity_id, version)
		);

		CREATE TABLE tia (
			id                BIGSERIAL   PRIMARY KEY,
			dpia_id           BIGINT      NOT NULL REFERENCES dpia(id),
			recipient_country TEXT        NOT NULL,
			safeguard_vi      TEXT        NOT NULL,
			created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Exec(context.Background(), `
			DROP TABLE tia CASCADE;
			DROP TABLE dpia CASCADE;
			DROP TABLE processing_activity CASCADE;
		`)
		pool.Close()
	})

	return NewService(NewPgRepo(pool)), pool
}

func countTIA(t *testing.T, db *pgxpool.Pool, activityID int64) int {
	var c int
	err := db.QueryRow(context.Background(), `
		SELECT count(*) FROM tia t
		JOIN dpia d ON d.id = t.dpia_id
		WHERE d.activity_id = $1
	`, activityID).Scan(&c)
	require.NoError(t, err)
	return c
}

func maxDPIAVersion(t *testing.T, db *pgxpool.Pool, activityID int64) int {
	var v int
	err := db.QueryRow(context.Background(), `
		SELECT max(version) FROM dpia WHERE activity_id = $1
	`, activityID).Scan(&v)
	require.NoError(t, err)
	return v
}

func basicActivity() ProcessingActivity {
	return ProcessingActivity{
		Name:           "Test Analytics",
		PurposeKey:     "analytics",
		DataCategories: []string{"behavior"},
		StartedAt:      time.Now(),
		CrossBorder:    false,
	}
}

func TestRegister_CrossBorderRequiresTIA(t *testing.T) {
	s, _ := setup(t)
	rc := "ID"
	a := ProcessingActivity{Name: "SEA analytics", CrossBorder: true, RecipientCountry: &rc, StartedAt: time.Now(), PurposeKey: "a", DataCategories: []string{}}
	_, err := s.RegisterActivity(context.Background(), a, DPIAInput{RiskLevel: "high"}) // thieu TIA
	require.ErrorIs(t, err, ErrTIARequired)
}

func TestRegister_CrossBorderWithTIA_CreatesTIA(t *testing.T) {
	s, db := setup(t)
	rc := "ID"
	a := ProcessingActivity{Name: "SEA analytics", CrossBorder: true, RecipientCountry: &rc, StartedAt: time.Now(), PurposeKey: "a", DataCategories: []string{}}
	in := DPIAInput{RiskLevel: "high", TIA: &TIAInput{RecipientCountry: "ID", Safeguard: "SCC + ma hoa"}}
	id, err := s.RegisterActivity(context.Background(), a, in)
	require.NoError(t, err)
	require.Equal(t, 1, countTIA(t, db, id))
}

func TestReview_CreatesNewVersion(t *testing.T) {
	s, db := setup(t)
	id, err := s.RegisterActivity(context.Background(), basicActivity(), DPIAInput{RiskLevel: "low"})
	require.NoError(t, err)
	require.NoError(t, s.ReviewDPIA(context.Background(), id, DPIAInput{RiskLevel: "medium"}))
	require.Equal(t, 2, maxDPIAVersion(t, db, id))
}
