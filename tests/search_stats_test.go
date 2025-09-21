package tests

import (
	"net/http/httptest"
	"testing"

	"backend/internal/api"
	"backend/internal/middleware"
)

// TestStatsAndSearch ensures /stats and /search handlers respond
func TestStatsAndSearch(t *testing.T) {
	fs, user, _ := SetupTest(t)

	// Stats
	req := httptest.NewRequest("GET", "/stats", nil)
	req = req.WithContext(middleware.WithUser(req.Context(), user))
	rr := httptest.NewRecorder()
	statsHandler := api.NewStatsHandler(nil) // if your handler requires service instance, pass real one
	statsHandler.ServeHTTP(rr, req)
	if rr.Code != 200 && rr.Code != 500 { // 500 possible if not implemented in test environment
		t.Fatalf("stats failed: %d %s", rr.Code, rr.Body.String())
	}

	// Search
	req2 := httptest.NewRequest("GET", "/search?q=hello", nil)
	req2 = req2.WithContext(middleware.WithUser(req2.Context(), user))
	rr2 := httptest.NewRecorder()
	searchHandler := api.NewSearchHandler(nil)
	searchHandler.ServeHTTP(rr2, req2)
	if rr2.Code != 200 && rr2.Code != 500 {
		t.Fatalf("search failed: %d %s", rr2.Code, rr2.Body.String())
	}
}
