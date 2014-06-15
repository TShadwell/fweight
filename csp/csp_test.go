package csp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorks(t *testing.T) {
	csp := ContentSecurityPolicy{
		Default: Self,
		Style:   Sources(Self, "*://fonts.googleapis.com/*"),
		Sandbox: Exceptions(AllowForms, AllowSameOrigin),
	}

	h := csp.Middleware(http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {}))

	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	var s string
	if s = w.Header().Get("Content-Security-Policy"); s == "" {
		t.Fatal("No CSP set!")
	}

}
