package server

import (
	"strings"

	"github.com/Abb133Se/httpServer/internal/utils"
)

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
	utils.Info("Initializing new router")
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
	method = strings.ToUpper(method)
	if _, ok := r.routes[path]; !ok {
		r.routes[path] = make(map[string]HandlerFunc)
	}
	r.routes[path][method] = handler
	utils.Debug("Registered route: %s %s", method, path)
}

// HandlePrefix registers a handler for all routes beginning with a prefix.
//
// Parameters:
//   - path:    Path prefix (e.g., "/files/").
//   - method:  HTTP method (e.g., "GET", "POST").
//   - handler: The handler function to execute for matching requests.
func (r *Router) HandlePrefix(path string, method string, handler HandlerFunc) {
	method = strings.ToUpper(method)
	if _, ok := r.prefixRoutes[path]; !ok {
		r.prefixRoutes[path] = make(map[string]HandlerFunc)
	}
	r.prefixRoutes[path][method] = handler
	utils.Debug("Registered prefix route: %s %s", method, path)
}

// Route dispatches a request to the appropriate handler.
//
// Matching priority:
//  1. Exact match
//  2. Prefix match
//  3. 404 Not Found if no match
//  4. 405 Method Not Allowed if method unsupported
//
// Parameters:
//   - req: The parsed HTTP request to route.
//
// Returns:
//   - Response: The response from the matched handler, or a generated error response.
func (r *Router) Route(req *Request) Response {
	method := strings.ToUpper(req.Method)
	if methods, ok := r.routes[req.Path]; ok {
		if handler, exists := methods[method]; exists {
			utils.Debug("Routing request: %s %s -> exact match", method, req.Path)
			return handler(req)
		}
		allow := GetAllowedMethods(methods)
		utils.Warn("Method not allowed for exact path: %s %s, allowed: %s", method, req.Path, allow)
		return MethodNotAllowedResponse(allow)
	}

	for prefix, methods := range r.prefixRoutes {
		if strings.HasPrefix(req.Path, prefix) {
			if handler, ok := methods[method]; ok {
				utils.Debug("Routing response: %s %s -> prefix match %s", method, req.Path, prefix)
				return handler(req)
			}
			allow := GetAllowedMethods(methods)
			utils.Warn("Methods not allowed on prefix: %s %s, allowed: %s", method, req.Path, allow)
			return MethodNotAllowedResponse(allow)
		}
	}

	utils.Warn("Route not found for method: %s %s", method, req.Path)
	return NotFoundResponse()
}

func GetAllowedMethods(methods map[string]HandlerFunc) string {
	var allowed []string
	for m := range methods {
		allowed = append(allowed, m)
	}
	return strings.Join(allowed, ", ")
}

// MethodNotAllowedResponse generates a 405 Method Not Allowed response.
func MethodNotAllowedResponse(allaw string) Response {
	return Response{
		Version: "HTTP/1.1",
		Status:  405,
		Reason:  "Method Not Allowed",
		Headers: map[string]string{
			"Content-Type": "text/plain",
			"Allow":        allaw,
		},
		Body: []byte("405 Method Not Allowed"),
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

// OptionsResponse generates an automatic 204 No Content OPTIONS response
// with the Allow header set.
func OptionsResponse(allow string) Response {
	utils.Info("Handling automatic OPTIONS response, Allow: %s", allow)
	return Response{
		Version: HTTPVersion,
		Status:  204, // No Content
		Reason:  "No Content",
		Headers: map[string]string{
			"Allow": allow,
		},
		Body: []byte(allow),
	}
}
