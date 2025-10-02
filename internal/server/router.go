package server

import "strings"

// HandlerFunc defines the function signature for all HTTP route handlers.
//
// A handler receives a parsed HTTP request and returns a Response struct.
// It must not write directly to the network connection.
//
// Example:
//
//	func helloHandler(req *Request) Response {
//	    return Response{
//	        Version: HTTPVersion,
//	        Status:  200,
//	        Reason:  "OK",
//	        Headers: map[string]string{"Content-Type": "text/plain"},
//	        Body:    []byte("Hello, World!"),
//	    }
//	}
type HandlerFunc func(req *Request) Response

// Router stores mappings of paths (exact or prefix) to handlers.
//
// It supports both exact matches ("/path") and prefix matches ("/files/…").
// Each path can have multiple handlers mapped by HTTP method.
type Router struct {
	routes       map[string]map[string]HandlerFunc // exact path → method → handler
	prefixRoutes map[string]map[string]HandlerFunc // prefix path → method → handler
}

// NewRouter creates and initializes a new Router.
//
// Returns:
//   - *Router: A pointer to a Router instance with no predefined routes.
func NewRouter() *Router {
	return &Router{
		routes:       map[string]map[string]HandlerFunc{},
		prefixRoutes: map[string]map[string]HandlerFunc{},
	}
}

// Handle registers a handler for an exact path and HTTP method.
//
// Parameters:
//   - path:    Exact match path (e.g., "/").
//   - method:  HTTP method (e.g., "GET", "POST").
//   - handler: The handler function to execute for this path+method.
func (r *Router) Handle(path string, method string, handler HandlerFunc) {
	if r.routes[path] == nil {
		r.routes[path] = make(map[string]HandlerFunc)
	}
	r.routes[path][method] = handler
}

// HandlePrefix registers a handler for all routes beginning with a prefix.
//
// Parameters:
//   - path:    Path prefix (e.g., "/files/").
//   - method:  HTTP method (e.g., "GET", "POST").
//   - handler: The handler function to execute for matching requests.
func (r *Router) HandlePrefix(path string, method string, handler HandlerFunc) {
	if r.prefixRoutes[path] == nil {
		r.prefixRoutes[path] = make(map[string]HandlerFunc)
	}
	r.prefixRoutes[path][method] = handler
}

// Route dispatches a request to the correct handler.
//
// Matching priority:
//   1. Exact match in routes.
//   2. Prefix match in prefixRoutes.
//   3. Returns a 404 Not Found if no match is found.
//   4. Returns a 405 Method Not Allowed if path matches but method does not.
//
// Parameters:
//   - req: The parsed HTTP request to route.
//
// Returns:
//   - Response: The response from the matched handler, or a generated error response.
func (r *Router) Route(req *Request) Response {
	if methods, ok := r.routes[req.Path]; ok {
		if h, ok := methods[req.Method]; ok {
			return h(req)
		}
		return MethodNotAllowedResponse()
	}

	for prefix, methods := range r.prefixRoutes {
		if strings.HasPrefix(req.Path, prefix) {
			if h, ok := methods[req.Method]; ok {
				return h(req)
			}
			return MethodNotAllowedResponse()
		}
	}
	return NotFoundResponse()
}

// MethodNotAllowedResponse generates a 405 Method Not Allowed response.
func MethodNotAllowedResponse() Response {
	return Response{
		Version: "HTTP/1.1",
		Status:  405,
		Reason:  "Method Not Allowed",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    []byte("405 Method Not Allowed"),
	}
}

// NotFoundResponse generates a 404 Not Found response.
func NotFoundResponse() Response {
	return Response{
		Version: "HTTP/1.1",
		Status:  404,
		Reason:  "Not Found",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    []byte("404 Not Found"),
	}
}
