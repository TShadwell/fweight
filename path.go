package fweight

import (
	"net/http"
	"strings"
)

var _ PathRouter = new(Path)
var _ Router = new(Path)

//Func PathIsEmpty accounts for the fact that paths like
// a/b///c or a/b// can exist, which would result in final
//strings of // or ///. These are assumed to be the same as
//their single-slashed counterparts.
func PathEmpty(path string) bool {
	if path == "" || strings.TrimLeft(path, "/") == "" {
		return true
	}
	return false
}

//Type Path is a Router which routes by URL path. Files or directories
//with empty names are not allowed. The empty name routes to the terminal
//Router, the one used if the path stops here.
//
//It should be noted that when RouteHTTP is called
//the PathRouter is followed to completion from the
//start to end of the URL, thus using two routers separately
//(i.e. separated by a different Router in the routing tree)
//causes two separate and independant operations on the path;
//a Path that routes a/b followed by one that routes a/b/c does
//not result in a route of a/b/a/c, instead resulting in a route of
//just a/b.
type Path map[string]Router

func (p Path) self() Path {
	if p == nil {
		p = make(Path)
	}
	return p
}

//Function AddChild adds a child PathRouter which is used
//as the next name in the path.
//
//A non PathRouter child will cause the Router to
//be returned to the caller of RouteHTTP, causing the
//path to terminate there if it is a Handler.
func (p Path) AddChild(pR Router, name string) Path {
	p[name] = pR
	return p
}

//Sets the Handler that is used when the path terminates here.
//This is the same as AddChild(r, ""); the empty string routes
//to here.
func (p Path) Handler(r Router) Path {
	return p.AddChild(r, "")
}

func isPathRouter(r Router) (b bool) {
	_, b = r.(PathRouter)
	return
}

//RouteHTTP traverses the tree of Paths until the end of the URL path
//is encountered, returning the terminal router or nil.
func (p Path) RouteHTTP(rq *http.Request) Router {
	remaining := strings.Trim(rq.URL.Path, "/")
	var pRouter Router
	for pRouter, remaining = p.Child(rq.URL.Path); isPathRouter(pRouter); pRouter, remaining = pRouter.(PathRouter).Child(remaining) {
	}
	return pRouter
}

//Function Child returns the next Router associated with the next
//'hop' in the path.
func (p Path) Child(subpath string) (n Router, remainingSubpath string) {
	//If strings.SplitN(subpath, "/", 3)'s length is two, then this
	//is the only item left in the path, and thus we must terminate here.
	if subpath == "" || subpath == "/" {
		return p[""], ""
	}

	//First, check if we have bound a handler for this whole
	//subpath (a/b/c)
	if pR, ok := p[subpath]; ok {
		return pR, ""
	}

	//Check if the next node is present
	splt := strings.SplitN(subpath, "/", 3)
	if pR, ok := p[splt[0]]; ok {
		return pR, splt[1]
	}

	//Check if we have a route that begins with the subpath
	for path, pR := range p {
		if strings.HasPrefix(subpath, path) {
			return pR, strings.TrimLeft(subpath, path)
		}
	}

	//Not Found.
	return nil, subpath
}
