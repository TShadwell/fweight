package route

import (
	"log"
	"net/http"
	Path "path"
	"reflect"
	"strings"
	"github.com/TShadwell/fweight"
)

//A PathingRouter is part of the filepath, and can take part
//in the descent of the filepath trie.
type PathingRouter interface {
	Child(subpath string) (n fweight.Router, remainingSubpath string)
}

var (
	_ PathingRouter  = new(PathRouter)
	_ fweight.Router = new(PathRouter)
)

//IgnoreExtensions is a PathRouter that ignores
//the extensions of folders and files when matching
//request URIs.
type IgnoreExtensions PathRouter

func (i IgnoreExtensions) underlying() PathRouter {
	return PathRouter(i)
}

//Function Child is provided by all types implementing the PathingRouter
//interface.
func (i IgnoreExtensions) Child(subpath string) (fweight.Router, string) {
	return i.underlying().ChildProcess(
		subpath,
		func(s string) string {
			ext := Path.Ext(s)
			return strings.TrimSuffix(s, ext)
		},
	)
}

func (i IgnoreExtensions) RouteHTTP(r *http.Request) fweight.Router {
	return PathingRouterRouteHTTP(i, r)
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
type PathRouter map[string]fweight.Router

func (p PathRouter) self() PathRouter {
	if p == nil {
		p = make(PathRouter)
	}
	return p
}

//Function AddChild adds a child PathRouter which is used
//as the next name in the path.
//
//A non PathRouter child will cause the Router to
//be returned to the caller of RouteHTTP, causing the
//path to terminate there if it is a Handler.
func (p PathRouter) AddChild(pR fweight.Router, name string) PathRouter {
	p[name] = pR
	return p
}

//Sets the Handler that is used when the path terminates here.
//This is the same as AddChild(r, ""); the empty string routes
//to here.
func (p PathRouter) Handler(r fweight.Router) PathRouter {
	return p.AddChild(r, "")
}

func isPathRouter(r fweight.Router) (b bool) {
	_, b = r.(PathRouter)
	return
}

//Performs RouteHTTP on a trie of PathingRouters.
func PathingRouterRouteHTTP(p PathingRouter, rq *http.Request) fweight.Router {
	/*
		Now like SubdomainRouter!
	*/
	var (
		currentPathingRouter PathingRouter = p
		path                 string        = strings.Trim(rq.URL.Path, "/")
		currentRouter        fweight.Router
	)

	//fix paths
	path = Path.Clean(path)

	//We break spec a bit to allow direcories called "."
	if path == "." {
		path = ""
	}
	for {
		currentRouter, path = currentPathingRouter.Child(path)

		var ok bool
		if currentPathingRouter, ok = currentRouter.(PathingRouter); !ok {
			if debug {
				log.Printf("[-] Path routing terminated at %v [%v] and pathRouter %v [%v]", currentRouter, reflect.TypeOf(currentRouter), currentPathingRouter, reflect.TypeOf(currentPathingRouter))
			}
			break
		}

	}

	return currentRouter
}

//Type AnyFile is a PathingRouter that accepts one file
//in the path, no matter what it is.
type AnyFile struct {
	fweight.Router
}

func (a AnyFile) RouteHTTP(rq *http.Request) fweight.Router {
	return PathingRouterRouteHTTP(a, rq)
}

func (a AnyFile) Child(subpath string) (n fweight.Router, remainingSubpath string) {
	return a.Router, subpath[strings.IndexRune(subpath, '/')+1:]
}

//RouteHTTP traverses the tree of Paths until the end of the URL path
//is encountered, returning the terminal router or nil.
func (p PathRouter) RouteHTTP(rq *http.Request) fweight.Router {
	return PathingRouterRouteHTTP(p, rq)
}

//Function Child is provided by all types that implement the PathingRouter interface.
func (p PathRouter) Child(subpath string) (n fweight.Router, remainingSubpath string) {
	return p.ChildProcess(subpath, nil)
}

//Function Child returns the next Router associated with the next
//'hop' in the path.
func (p PathRouter) ChildProcess(subpath string, process func(string) string) (n fweight.Router, remainingSubpath string) {

	if process == nil {
		process = func(s string) string {
			return s
		}
	}

	if debug {
		log.Printf("[?] Currently in path %+v, with requested path %+q \n", p, subpath)
	}

	//strip leading slashes
	subpath = strings.TrimLeft(subpath, "/")

	//If strings.SplitN(subpath, "/", 3)'s length is two, then this
	//is the only item left in the path, and thus we must terminate here.
	if subpath == "" || subpath == "/" {
		if debug {
			log.Println("[?] Routing into current level (path is empty).")
		}
		return p[""], ""
	}

	processedSubpath := process(subpath)

	//First, check if we have bound a handler for this whole
	//subpath (a/b/c)
	if pR, ok := p[processedSubpath]; ok {
		return pR, ""
	} else if debug {
		log.Printf("[?] %v not wholly in %+v\n", subpath, p)
	}

	//Check if the next node is present
	splt := strings.SplitN(subpath, "/", 2)
	if len(splt) > 1 {
		if pR, ok := p[process(splt[0])]; ok {
			if debug {
				log.Printf("[?] Routed down into %v remaining string %+q.\n", reflect.TypeOf(pR), splt[1])
			}
			return pR, splt[1]
		}
	} else if debug {
		log.Printf("[?] %+v too short.\n", splt)
	}

	//Check if we have a route that begins with the subpath
	/*
		for path, pR := range p {
			if path != "" && strings.HasPrefix(processedSubpath, path) {
				if debug {
					log.Printf(
						"[?] Routed into %v - %+q is a prefix of %+q",
						pR,
						path,
						subpath,
					)
				}
				return pR, strings.TrimLeft(subpath, path)
			} else if debug {
				log.Printf(
					"[?] Did not reoute %+q is not a prefix of %+q",
					path,
					subpath,
				)
			}
		}
		//I actually have no idea what this does
	*/

	//Not Found.
	return nil, subpath
}
