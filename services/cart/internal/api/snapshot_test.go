package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"shopass/services/cart/internal/cart"
)

func TestSnapshotHandler_CreateSnapshot_Auth(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	repo := cart.NewSnapshotRepo(db)
	handler := NewSnapshotHandler(repo)

	// Mock DB behavior
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM cart_snapshot").WillReturnError(sql.ErrNoRows) // skip idempotent
	mock.ExpectQuery("INSERT INTO cart_snapshot").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectPrepare("INSERT INTO cart_item")
	mock.ExpectExec("INSERT INTO cart_item").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	payload := map[string]interface{}{
		"platform_id":  1,
		"snapshot_ref": uuid.New().String(),
		"items": []map[string]interface{}{
			{"qty": 1, "unit_price": 50000},
		},
		"cookie": "mysecretcookie", // This should be dropped naturally by strict unmarshal
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/v1/cart/snapshot", bytes.NewReader(body))
	// Gateway injects the authenticated user via header (TASK-CART-002).
	req.Header.Set("X-User-Id", "999")

	rr := httptest.NewRecorder()
	handler.CreateSnapshot(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var snap cart.CartSnapshot
	err := json.Unmarshal(rr.Body.Bytes(), &snap)
	assert.NoError(t, err)

	// Check if user_id is from context, not payload
	assert.Equal(t, int64(999), snap.UserID)
}

func TestSnapshotHandler_CreateSnapshot_InvalidPayload(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	repo := cart.NewSnapshotRepo(db)
	handler := NewSnapshotHandler(repo)

	payload := map[string]interface{}{
		"platform_id":  1,
		"snapshot_ref": uuid.New().String(),
		"items": []map[string]interface{}{
			{"qty": 0, "unit_price": 50000}, // qty = 0 is invalid
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/v1/cart/snapshot", bytes.NewReader(body))
	req.Header.Set("X-User-Id", "999")
	rr := httptest.NewRecorder()
	handler.CreateSnapshot(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
