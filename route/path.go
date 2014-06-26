package route

import (
	"log"
	"net/http"
	Pathp "path"
	"reflect"
	"strings"
)

func isSpecial(s string) bool {
	switch s {
	case "&", "":
		return true
	default:
		return false
	}
}

//A PathingRouter is part of the filepath, and can take part
//in the descent of the filepath trie.
type PathingRouter interface {
	Child(subpath string) (n Router, remainingSubpath string)
}

var (
	_ PathingRouter = new(Path)
	_ Router        = new(Path)
)

//NoExtPath is a PathRouter that ignores
//the extensions of folders and files when matching
//request URIs.
type NoExtPath Path

func (i NoExtPath) underlying() Path {
	return Path(i)
}

//Function Child is provided by all types implementing the PathingRouter
//interface.
func (i NoExtPath) Child(subpath string) (Router, string) {
	return i.underlying().ChildProcess(
		subpath,
		func(s string) string {
			ext := Pathp.Ext(s)
			return strings.TrimSuffix(s, ext)
		},
	)
}

func (i NoExtPath) RouteHTTP(r *http.Request) Router {
	return PathRouteHTTP(i, r)
}

//Type Path is a Router which routes by URL path. Files or directories
//with empty names are not allowed. The empty name routes to the terminal
//Router, the one used if the path stops here, and the ampersand ("&") path
//swallows the next file segment in the path, regardless of its contents.
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
	_, b = r.(Path)
	return
}

//Performs RouteHTTP on a trie of PathingRouters.
func PathRouteHTTP(p PathingRouter, rq *http.Request) Router {
	/*
		Now like SubdomainRouter!
	*/
	var (
		currentPathingRouter PathingRouter = p
		path                 string        = strings.Trim(rq.URL.Path, "/")
		currentRouter        Router
	)

	//fix paths
	path = Pathp.Clean(path)

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

//RouteHTTP traverses the tree of Paths until the end of the URL path
//is encountered, returning the terminal router or nil.
func (p Path) RouteHTTP(rq *http.Request) Router {
	return PathRouteHTTP(p, rq)
}

//Function Child is provided by all types that implement the PathingRouter interface.
func (p Path) Child(subpath string) (n Router, remainingSubpath string) {
	return p.ChildProcess(subpath, nil)
}

func swallowOne(subpath string) (subsectpath, swallowed string) {
	slashPos := strings.Index(subpath, "/")
	if slashPos == -1 {
		swallowed = subpath
		return
	}

	slashPos += 1
	if slashPos > len(subpath) {
		swallowed = subpath
		return
	}
	subsectpath, swallowed = subpath[slashPos:], subpath[:slashPos-1]
	return
}

//Function Child returns the next Router associated with the next
//'hop' in the path.
func (p Path) ChildProcess(subpath string, process func(string) string) (n Router, remainingSubpath string) {

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

	remaining, popped := swallowOne(subpath)

	if debug {
		log.Printf("If we chop one off, we get %s, %s\n", remaining, popped)
	}

	if pathRouter, ok := p[process(popped)]; ok {
		return pathRouter, remaining
	} else if debug {
		log.Printf("%s was not present.\n", process(popped))
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

	if p["&"] != nil {
		return p["&"], remaining
	} else if debug {
		log.Printf("[?] No ampersand present in Path, no swallow.")
	}

	//Not Found.
	return nil, subpath
}

//Returns s up until a char in terminators, or the whole string.
func TrimPast(s, terminators string) string {
	pos := strings.IndexAny(s, terminators)
	if pos == -1 {
		//to take the whole string
		pos = len(s)
	}

	return s[:pos]
}

//Returns s up until r, or the whole string.
func TrimPastRune(s string, r rune) string {
	pos := strings.IndexRune(s, r)
	if pos == -1 {
		//whole string
		pos = len(s)
	}
	return s[:pos]
}

//Function Ampersand is used to retrieve the URL part that was swallowed
//by an ampersand path. To achieve this, `prefix` is trimmed from the
//beginning of the url, and the resulting string is returned up
//until an instance of a char in `terminators` or the end of the string.
func Ampersand(url, prefix, terminators string) string {
	return TrimPast(strings.TrimPrefix(url, prefix), terminators)
}

//Function PartN returns the Nth part of a URL, separated by '/'.
//A leading slash ('/usr') is ignored.
//If there are too few parts, empty string is returned.
func PartN(url string, n int) string {
	if url[0] == '/' {
		if len(url) < 2 {
			return ""
		}
		url = url[1:]
	}
	for i, ch := range url {
		if ch == '/' {
			n--
		}

		if n == 0 {
			return TrimPast(
				strings.TrimLeft(url[i+1:], "/"),
				"/",
			)
		}
	}
	return ""
}

//Function Depth returns the depth in levels of a URL, separated by '/'.
func Depth(url string) int {
	s := strings.Trim(url, "/")
	if s == "" {
		return 0
	}
	return strings.Count(s, "/") + 1
}
