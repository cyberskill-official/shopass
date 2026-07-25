package dpia

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRepo struct {
	db *pgxpool.Pool
}

func NewPgRepo(db *pgxpool.Pool) repo {
	return &pgRepo{db: db}
}

func (r *pgRepo) createWithDPIA(ctx context.Context, a ProcessingActivity, in DPIAInput) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var activityID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO processing_activity (name, purpose_key, data_categories, started_at, cross_border, recipient_country)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, a.Name, a.PurposeKey, a.DataCategories, a.StartedAt, a.CrossBorder, a.RecipientCountry).Scan(&activityID)
	if err != nil {
		return 0, fmt.Errorf("insert activity: %w", err)
	}

	var dpiaID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO dpia (activity_id, version, risk_level, mitigation_vi)
		VALUES ($1, 1, $2, $3)
		RETURNING id
	`, activityID, in.RiskLevel, in.MitigationVi).Scan(&dpiaID)
	if err != nil {
		return 0, fmt.Errorf("insert dpia: %w", err)
	}

	if a.CrossBorder && in.TIA != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO tia (dpia_id, recipient_country, safeguard_vi)
			VALUES ($1, $2, $3)
		`, dpiaID, in.TIA.RecipientCountry, in.TIA.Safeguard)
		if err != nil {
			return 0, fmt.Errorf("insert tia: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return activityID, nil
}

func (r *pgRepo) createDPIAVersion(ctx context.Context, activityID int64, in DPIAInput) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var lastVersion int
	var oldStatus string
	var oldFiledAt *time.Time

	err = tx.QueryRow(ctx, `
		SELECT version, status, filed_at 
		FROM dpia 
		WHERE activity_id = $1 
		ORDER BY version DESC 
		LIMIT 1
	`, activityID).Scan(&lastVersion, &oldStatus, &oldFiledAt)
	if err != nil {
		return fmt.Errorf("get last version: %w", err)
	}

	var dpiaID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO dpia (activity_id, version, risk_level, mitigation_vi, status, filed_at, last_reviewed_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		RETURNING id
	`, activityID, lastVersion+1, in.RiskLevel, in.MitigationVi, oldStatus, oldFiledAt).Scan(&dpiaID)
	if err != nil {
		return fmt.Errorf("insert new dpia version: %w", err)
	}

	if in.TIA != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO tia (dpia_id, recipient_country, safeguard_vi)
			VALUES ($1, $2, $3)
		`, dpiaID, in.TIA.RecipientCountry, in.TIA.Safeguard)
		if err != nil {
			return fmt.Errorf("insert new tia version: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *pgRepo) markFiled(ctx context.Context, dpiaID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE dpia 
		SET filed_at = now() 
		WHERE id = $1
	`, dpiaID)
	return err
}

func (r *pgRepo) overdue(ctx context.Context) ([]ActivityStatus, error) {
	// Let's implement the DB logic to list overdues as per the spec query.
	// We will compute status in Go to reuse `Status` func, so we just fetch all latest DPIAs.
	rows, err := r.db.Query(ctx, `
		SELECT pa.name, pa.started_at, d.version, d.filed_at, d.last_reviewed_at
		FROM processing_activity pa
		JOIN LATERAL (
		  SELECT version, filed_at, last_reviewed_at 
		  FROM dpia 
		  WHERE activity_id = pa.id 
		  ORDER BY version DESC 
		  LIMIT 1
		) d ON true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ActivityStatus
	now := time.Now()

	for rows.Next() {
		var name string
		var startedAt time.Time
		var version int
		var filedAt, lastReviewedAt *time.Time
		if err := rows.Scan(&name, &startedAt, &version, &filedAt, &lastReviewedAt); err != nil {
			return nil, err
		}

		status := Status(
			ProcessingActivity{StartedAt: startedAt},
			DPIA{FiledAt: filedAt, LastReviewedAt: lastReviewedAt},
			now,
		)
		if status == "overdue" || status == "review_overdue" {
			res = append(res, ActivityStatus{
				Name:      name,
				StartedAt: startedAt,
				Version:   version,
				Status:    status,
			})
		}
	}
	return res, rows.Err()
}

func (r *pgRepo) report(ctx context.Context) ([]ActivityStatus, error) {
	rows, err := r.db.Query(ctx, `
		SELECT pa.name, pa.started_at, d.version, d.filed_at, d.last_reviewed_at
		FROM processing_activity pa
		JOIN LATERAL (
		  SELECT version, filed_at, last_reviewed_at 
		  FROM dpia 
		  WHERE activity_id = pa.id 
		  ORDER BY version DESC 
		  LIMIT 1
		) d ON true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ActivityStatus
	now := time.Now()

	for rows.Next() {
		var name string
		var startedAt time.Time
		var version int
		var filedAt, lastReviewedAt *time.Time
		if err := rows.Scan(&name, &startedAt, &version, &filedAt, &lastReviewedAt); err != nil {
			return nil, err
		}

		status := Status(
			ProcessingActivity{StartedAt: startedAt},
			DPIA{FiledAt: filedAt, LastReviewedAt: lastReviewedAt},
			now,
		)
		res = append(res, ActivityStatus{
			Name:      name,
			StartedAt: startedAt,
			Version:   version,
			Status:    status,
		})
	}
	return res, rows.Err()
}
