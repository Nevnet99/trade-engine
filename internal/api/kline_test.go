package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nevnet99/trade-engine/internal/store"
	"github.com/Nevnet99/trade-engine/internal/testutils"
)

func TestHandleGetKlines(t *testing.T) {
	tx := testutils.SetupTestDB(t)
	storage := store.NewStorage(tx)
	s := NewServer(storage)

	tests := []struct {
		name       string
		target     string
		wantStatus int
	}{
		{
			name:       "Happy Path",
			target:     "/kline?symbol=BTC-USD&interval=1m",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Missing Symbol",
			target:     "/kline?interval=1m",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Missing Interval",
			target:     "/kline?symbol=BTC-USD",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid Interval (Not in allowlist)",
			target:     "/kline?symbol=BTC-USD&interval=2m", // 2m is not supported
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid Limit (Non-numeric)",
			target:     "/kline?symbol=BTC-USD&interval=1m&limit=abc",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			w := httptest.NewRecorder()

			s.HandleGetKlines(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status Code: got %d, want %d. Body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}
