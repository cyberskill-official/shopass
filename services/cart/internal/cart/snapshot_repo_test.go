package cart

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSnapshotRepo_InsertSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewSnapshotRepo(db)
	ctx := context.Background()

	ref := uuid.New()
	snap := &CartSnapshot{
		UserID:      1,
		PlatformID:  1,
		SnapshotRef: ref,
		CapturedAt:  time.Now(),
		Items: []CartItem{
			{Qty: 2, UnitPrice: 150000},
		},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM cart_snapshot WHERE user_id = \\$1 AND snapshot_ref = \\$2").
		WithArgs(snap.UserID, snap.SnapshotRef).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery("INSERT INTO cart_snapshot").
		WithArgs(snap.UserID, snap.PlatformID, snap.SnapshotRef, snap.CapturedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))

	mock.ExpectPrepare("INSERT INTO cart_item")
	mock.ExpectExec("INSERT INTO cart_item").
		WithArgs(100, nil, nil, nil, snap.Items[0].Qty, snap.Items[0].UnitPrice).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.InsertSnapshot(ctx, snap)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), snap.ID)
	assert.Equal(t, int64(100), snap.Items[0].CartSnapshotID)
}

func TestSnapshotRepo_InsertSnapshot_Idempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewSnapshotRepo(db)
	ctx := context.Background()

	ref := uuid.New()
	snap := &CartSnapshot{
		UserID:      1,
		PlatformID:  1,
		SnapshotRef: ref,
		CapturedAt:  time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM cart_snapshot WHERE user_id = \\$1 AND snapshot_ref = \\$2").
		WithArgs(snap.UserID, snap.SnapshotRef).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(200)) // Already exists

	mock.ExpectCommit()

	err = repo.InsertSnapshot(ctx, snap)
	assert.NoError(t, err)
	assert.Equal(t, int64(200), snap.ID)
}
