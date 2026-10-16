package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetLiterature_minIntervalBetweenRequests(t *testing.T) {
	literatureReqLastAt = time.Time{}

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	c := NewHTTPClient()
	ctx := context.Background()

	start := time.Now()
	if _, err := c.GetLiterature(ctx, srv.URL, nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetLiterature(ctx, srv.URL, nil, nil); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if hits != 2 {
		t.Fatalf("hits %d", hits)
	}
	if elapsed < LiteratureOutboundMinInterval {
		t.Fatalf("elapsed %v want >= %v", elapsed, LiteratureOutboundMinInterval)
	}
}
