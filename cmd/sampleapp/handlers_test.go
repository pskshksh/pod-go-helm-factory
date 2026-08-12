package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRoutes confirms the probe and root endpoints answer, and unknown paths
// 404 through the catch-all.
func TestRoutes(t *testing.T) {
	cases := []struct {
		path string
		code int
	}{
		{"/healthz", http.StatusOK},
		{"/readyz", http.StatusOK},
		{"/", http.StatusOK},
		{"/nope", http.StatusNotFound},
	}

	handler := routes()
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, c.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != c.code {
				t.Errorf("GET %s = %d, want %d", c.path, rec.Code, c.code)
			}
		})
	}
}
