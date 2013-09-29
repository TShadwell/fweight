/*
	Package fweight impliments a http.Handler with nice syntax and interfaces
	that can stand high-performence using tries for routing, and extends the
	http.ResponseWriter interface with some useful tools.

	For now, it expects to be accessed with a hostname and not through an IP.
	Accessing it through an IP will result in NotFound.

		http.Handle(
			":8080",
			new(Server).
			Domain(
				AnyDomainThen(
					new(Subdomain).Domain(
						"test",
						new(PathRouter).
						Get(helloHandler).
						Put(putHandler),
					),
				),
			),
		)

	The concept of fweight is to provide a modular framework
	that is backward compatible with typical http.Handlers.

*/
package fweight
