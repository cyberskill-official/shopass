package consent

import (
	"context"
	"database/sql"
)

type Repo interface {
	latest(ctx context.Context, userID int64, purposeKey string) (ConsentRecord, error)
	append(ctx context.Context, rec ConsentRecord) error
	effectiveVersion(ctx context.Context, purposeKey string) (int32, error)
	history(ctx context.Context, userID int64, purposeKey string) ([]ConsentRecord, error)
}

type pgRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) Repo {
	return &pgRepo{db: db}
}

func (r *pgRepo) latest(ctx context.Context, userID int64, purposeKey string) (ConsentRecord, error) {
	query := `
		SELECT id, user_id, purpose_key, policy_version, granted, source, ts, ip, user_agent
		FROM consent_record
		WHERE user_id = $1 AND purpose_key = $2
		ORDER BY ts DESC
		LIMIT 1
	`
	var rec ConsentRecord
	var ip, ua sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID, purposeKey).Scan(
		&rec.ID, &rec.UserID, &rec.PurposeKey, &rec.PolicyVersion,
		&rec.Granted, &rec.Source, &rec.TS, &ip, &ua,
	)
	if err != nil {
		return rec, err
	}
	// parsing IP/UA omitted for brevity in this simple mapping
	return rec, nil
}

func (r *pgRepo) append(ctx context.Context, rec ConsentRecord) error {
	query := `
		INSERT INTO consent_record (user_id, purpose_key, policy_version, granted, source, ts, ip, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	var ipStr, uaStr interface{}
	if rec.IP != nil {
		ipStr = rec.IP.String()
	}
	if rec.UserAgent != nil {
		uaStr = *rec.UserAgent
	}
	// Note: in actual implementation, we might let postgres assign ts if it's zero,
	// but passing it explicitely works too if it's set. If it's zero we can let pg set it.
	// For simplicity, let's just pass what's in rec and assume caller sets TS or we use `NOW()`
	if rec.TS.IsZero() {
		query = `
			INSERT INTO consent_record (user_id, purpose_key, policy_version, granted, source, ip, user_agent)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		_, err := r.db.ExecContext(ctx, query, rec.UserID, rec.PurposeKey, rec.PolicyVersion, rec.Granted, rec.Source, ipStr, uaStr)
		return err
	}

	_, err := r.db.ExecContext(ctx, query, rec.UserID, rec.PurposeKey, rec.PolicyVersion, rec.Granted, rec.Source, rec.TS, ipStr, uaStr)
	return err
}

func (r *pgRepo) effectiveVersion(ctx context.Context, purposeKey string) (int32, error) {
	query := `
		SELECT version
		FROM consent_policy
		WHERE purpose_key = $1 AND effective_from <= now()
		ORDER BY effective_from DESC, version DESC
		LIMIT 1
	`
	var version int32
	err := r.db.QueryRowContext(ctx, query, purposeKey).Scan(&version)
	return version, err
}

func (r *pgRepo) history(ctx context.Context, userID int64, purposeKey string) ([]ConsentRecord, error) {
	query := `
		SELECT id, user_id, purpose_key, policy_version, granted, source, ts, ip, user_agent
		FROM consent_record
		WHERE user_id = $1 AND purpose_key = $2
		ORDER BY ts ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, purposeKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ConsentRecord
	for rows.Next() {
		var rec ConsentRecord
		var ip, ua sql.NullString
		if err := rows.Scan(
			&rec.ID, &rec.UserID, &rec.PurposeKey, &rec.PolicyVersion,
			&rec.Granted, &rec.Source, &rec.TS, &ip, &ua,
		); err != nil {
			return nil, err
		}
		res = append(res, rec)
	}
	return res, rows.Err()
}
