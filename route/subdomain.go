package route

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"reflect"
	"strings"
)

/*
	Internally, the domain system is based on the DomainRouter interface.
	A DomainRouter will check an assertion that the Router that is recieved
	from the Subdomain function is a DomainRouter. If is, the RouteHTTP
	function continues down the trie until this is not the case.
*/
type DomainRouter interface {
	Subdomain(subpath string) (s Router, remainingDomain string)
	Router
}

var _ DomainRouter = &Subdomain{}
var _ Router = &Subdomain{}

//function removeLevel removes the highest level domain from a domain name
func popLevel(domain string) (newDomain, oldLevel string) {
	var lastdot uint
	for i, v := range domain {
		if v == '.' {
			lastdot = uint(i)
		}
	}
	return domain[:lastdot], domain[lastdot+1:]
}

/*
	Subdomain implements the DomainRouter interface and
	is used to route requests to subdomain trees and the paths
	below them.

	The empty subdomain ("") is used when the route terminates here.
*/
type Subdomain map[string]Router

const termHere = ""

func (s Subdomain) Here(r Router) {
	s[termHere] = r
}

func (s Subdomain) here() Router {
	return s[termHere]
}

func removeSubdomain(subd, path string) (s string) {
	s = strings.TrimSuffix(path, subd)
	if len(s) > 0 {
		s = strings.Trim(s, ".")
	}

	return
}

// isSubdomin returns the value of the assertion
// `if r implements SubdomainRouter`. If the assertion is true,
// sd is set to the DomainRouter of r.
func isSubdomain(r Router, sd DomainRouter) (b bool) {
	sd, b = r.(DomainRouter)
	return
}

/*
	RouteHTTP completes a route down a series of DomainRouters.
	Once the router is no longer a DomainRouter, it is returned.

*/
func (s Subdomain) RouteHTTP(rq *http.Request) Router {
	var (
		currentSubdomain DomainRouter = s
		domain           string       = rq.Host
		currentRouter    Router
	)
	//Remove port
	if subp := strings.SplitN(domain, ":", 2); len(subp) > 1 {
		domain = subp[0]
	}

	//fix odd domains (x.com..x)
	domain = strings.Replace(path.Clean(strings.Replace(domain, ".", "/", -1)), "/", ".", -1)

	for {
		if debug {
			log.Printf("Route is now %+q\n", domain)
		}
		currentRouter, domain = currentSubdomain.Subdomain(domain)

		var ok bool
		if currentSubdomain, ok = currentRouter.(DomainRouter); !ok {
			if debug {
				log.Println("Subdomain routing terminated at:", reflect.TypeOf(currentRouter), reflect.TypeOf(currentSubdomain))
			}
			break
		}
	}

	return currentRouter
}

func (s Subdomain) String() (o string) {
	o = "["
	for k, v := range s {
		o += fmt.Sprintf("%+q -> %v\n", k, reflect.ValueOf(v))
	}
	return o[:len(o)-1] + "]"
}

func debRoute(ty, message string, v interface{}) {
	var targetS string
	if k, ok := v.(Subdomain); ok {
		targetS = k.String()
	} else {
		targetS = reflect.ValueOf(v).String()
	}

	fmt.Printf("%+q, %+q, %+q\n", ty, message, targetS)
}

//Function Subdomain is provided by all types implementing the
//SubdomainRouter interface.
func (s Subdomain) Subdomain(subpath string) (Router, string) {

	//Check if we have bound a handler for the entire remaining route.
	if sD, ok := s[subpath]; ok {
		if debug {
			debRoute(subpath, "represents whole path in", sD)
		}
		//Nothing left.
		return sD, ""
	}

	//Check if the next node is present
	var cSubpath, cLevel string
	cSubpath, cLevel = popLevel(subpath)
	//If the subpath is "empty", then we return this Subdomain's PathRouter
	if rT, ok := s[cLevel]; ok {
		if debug {
			debRoute(cLevel, "routes to ", rT)
		}
		return rT, cSubpath
	}

	//If the requested domain is the suffix of the current domain
	//strip off that component as per its map.
	for subDomain, router := range s {
		if subDomain != termHere && strings.HasSuffix(subpath, subDomain) {
			if debug {
				debRoute(subDomain, "is a suffix of", subpath)
			}
			return router, removeSubdomain(subDomain, subpath)
		}
	}

	if debug {
		debRoute(subpath, "-- none matched, 404", nil)
	}
	return nil, subpath
}

func (s Subdomain) Domain(name string, r Router) Subdomain {
	if s == nil {
		s = map[string]Router{
			name: r,
		}
	} else {
		s[name] = r
	}
	return s
}
