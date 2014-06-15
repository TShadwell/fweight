package csp

import (
	"fmt"
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

func ExampleContentSecurityPolicy_Middleware() {
	csp := ContentSecurityPolicy{
		Default: Self,
		Style: Sources(
			Self,
			"fonts.googleapis.com",
			UnsafeInline,
		),
		Script: Sources(
			Self,
			UnsafeInline,
		),
		Font: Sources(
			Self,
			Data,
			"themes.googleusercontent.com",
		),
		Sandbox: Exceptions(
			AllowForms,
			AllowSameOrigin,
			AllowScripts,
		),
	}

	h := csp.Middleware(http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {}))

	req, err := http.NewRequest("GET", "http://example.com/foo", nil)
	if err != nil {
		panic(err)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	fmt.Print(w.Header().Get("Content-Security-Policy"))
	// Output: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' fonts.googleapis.com 'unsafe-inline'; font-src 'self' data: themes.googleusercontent.com; sandbox allow-forms allow-same-origin allow-scripts
}
