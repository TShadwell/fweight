//Package csp implments Content-Security-Policy, a HTTP header
//designed to mitigate XSS attacks.
package csp

import (
	"github.com/TShadwell/fweight/route"
	"net/http"
	"reflect"
	"strings"
)

//A SandboxExceptionList is a space-separated list of identifiers
//specifying which exceptions to make to the sandbox directive.
type SandboxExceptionList string

const (
	//Allow form submission
	AllowForms SandboxExceptionList = "allow-forms"
	//Read raw mouse movement--
	//https://dvcs.w3.org/hg/pointerlock/raw-file/default/index.html
	AllowPointerLock SandboxExceptionList = "allow-pointer-lock"
	//Allow creation of 'auxillary browsing contexts', AKA popups.
	AllowPopups SandboxExceptionList = "allow-popups"
	//Allow scripts to access content on the same origin
	AllowSameOrigin SandboxExceptionList = "allow-same-origin"
	//Allow scripts to run
	AllowScripts SandboxExceptionList = "allow-scripts"
	//Allow this context to manipulate its parent context (window, frame).
	//http://www.whatwg.org/specs/web-apps/current-work/multipage/origin-0.html#sandboxed-top-level-navigation-browsing-context-flag
	AllowTopNavigation SandboxExceptionList = "allow-top-navigation"
)

//A sourcelist is a space-separated list of identifiers
//specifying which sources are acceptible.
type SourceList string

const (
	Any SourceList = "*"
	//Specifies that no sources are acceptible
	None SourceList = "'none'"
	//Same origin (same scheme, host, and port)
	Self SourceList = "'self'"
	//Via HTTPS
	HTTPS SourceList = "https:"
	//Via data
	Data SourceList = "data:"
	//Allow use of inline source elements (onclick, attribute, script tag bodies, onload;
	//depends on the directive it is part of).
	UnsafeInline SourceList = "'unsafe-inline'"
	//Allows unsafe dynamic code evaluation such as JavaScript eval()
	UnsafeEval SourceList = "'unsafe-eval'"
)

//Joins the sources in 's' with spaces.
func Sources(s ...SourceList) (so SourceList) {
	for _, v := range s {
		so += v + " "
	}
	so = so[:len(so)-1]
	return
}

//Joins the exceptions in 'e' with spaces.
func Exceptions(e ...SandboxExceptionList) (eo SandboxExceptionList) {
	for _, v := range e {
		eo += v + " "
	}
	eo = eo[:len(eo)-1]
	return
}

type ContentSecurityPolicy struct {
	//General valid rules for matching all loaded content
	Default SourceList "default-src"
	//Rules for loading scripts
	Script SourceList "script-src"
	//Rules for loading styles
	Style SourceList "style-src"
	//Rules for loading images
	Image SourceList "img-src"
	//Rules for AJAX, websockets and EventSource.
	//400 is emulated on failure.
	Connect SourceList "connect-src"
	//Rules for loading fonts
	Font SourceList "font-src"
	//Rules for loading <object>, <embed> and <applet>
	Object SourceList "object-src"
	//Rules for loading <audio> and <video>
	Media SourceList "media-src"
	//Rules for loading frames
	Frame SourceList "frame-src"

	//A series of strings representing what policies to ignore in the sandbox
	//for this resource.
	//To sandbox with no exception, set a non-empty value
	//with length zero.
	Sandbox SandboxExceptionList "sandbox"

	//Instructs the browser to POST reports of policy failures to this URI
	Report string "report-uri"
}

//Returns the Handler that would result from applying .Middleware to the given handler.
func (c ContentSecurityPolicy) RouteHandler(h http.Handler) route.Handler {
	return route.Handle(c.Middleware(h))
}

type cspHandler struct {
	policy  string
	handler http.Handler
}

func (c cspHandler) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	rw.Header().Set("Content-Security-Policy", c.policy)
	c.handler.ServeHTTP(rw, rq)
}

//Applies the Content Security Policy specified by 'c' to the http.Handler h.
func (c ContentSecurityPolicy) Middleware(h http.Handler) http.Handler {
	var csph cspHandler
	var directives []string
	v := reflect.ValueOf(c)
	t := v.Type()
	for i, nf := 0, t.NumField(); i < nf; i++ {
		f := t.Field(i)
		if f.Tag == "" {
			continue
		}

		sl := v.Field(i)
		if sl.Kind() != reflect.String {
			continue
		}

		st := sl.String()
		if st == "" {
			continue
		}

		directives = append(directives, string(f.Tag)+" "+st)
	}
	csph.policy = strings.Join(directives, "; ")

	if v, ok := h.(cspHandler); ok {
		csph.handler = v.handler
	} else {
		csph.handler = h
	}

	return csph
}
