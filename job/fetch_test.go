package job

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("first\nsecond\nthird\n"))
	}))
	defer server.Close()

	rows, err := fetch(server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
}

func TestFetchStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	rows, err := fetch(server.URL)
	if err == nil {
		t.Fatalf("expected error, got nil with rows: %v", rows)
	}
}
