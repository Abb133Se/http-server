package server

import (
	"regexp"
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

type Route struct {
	pattern   string
	method    string
	handler   HandlerFunc
	paramKeys []string       // for path parameters
	regex     *regexp.Regexp // compiled regex if it's a regex route
	isPrefix  bool
}

type Router struct {
	routes []*Route
	groups []*RouteGroup
}

type RouteGroup struct {
	prefix string
	routes []*Route
}

// NewRouter creates and initializes a new Router.
//
// Returns:
//   - *Router: A pointer to a Router instance with no predefined routes.
func NewRouter() *Router {
	utils.Info("Initializing new router")
	return &Router{
		routes: []*Route{},
		groups: []*RouteGroup{},
	}
}

// Handle registers a handler for an exact path and HTTP method.
//
// Parameters:
//   - path:    Exact match path (e.g., "/").
//   - method:  HTTP method (e.g., "GET", "POST").
//   - handler: The handler function to execute for this path+method.
func (r *Router) Handle(path, method string, handler HandlerFunc) {
	method = strings.ToUpper(method)
	route := &Route{
		pattern: path,
		method:  method,
		handler: handler,
	}
	r.routes = append(r.routes, route)
	utils.Debug("Registered route: %s %s", method, path)
}

// HandlePrefix registers a handler for all routes beginning with a prefix.
//
// Parameters:
//   - path:    Path prefix (e.g., "/files/").
//   - method:  HTTP method (e.g., "GET", "POST").
//   - handler: The handler function to execute for matching requests.
// func (r *Router) HandlePrefix(path string, method string, handler HandlerFunc) {
// 	method = strings.ToUpper(method)
// 	if _, ok := r.prefixRoutes[path]; !ok {
// 		r.prefixRoutes[path] = make(map[string]HandlerFunc)
// 	}
// 	r.prefixRoutes[path][method] = handler
// 	utils.Debug("Registered prefix route: %s %s", method, path)
// }

func (r *Router) Group(prefix string) *RouteGroup {
	group := &RouteGroup{prefix: prefix}
	r.groups = append(r.groups, group)
	return group
}

func (r *Router) HandleRegex(pattern string, handler HandlerFunc) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	route := &Route{
		pattern: pattern,
		handler: handler,
		regex:   re,
	}
	r.routes = append(r.routes, route)
	utils.Debug("Registered regex route: %s", pattern)
	return nil
}

func (g *RouteGroup) Handle(path string, handler HandlerFunc) {
	fullPath := g.prefix + path
	route := &Route{
		pattern: fullPath,
		handler: handler,
	}
	g.routes = append(g.routes, route)
	utils.Debug("Registered grouped route: %s", fullPath)
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
	for _, route := range r.routes {
		if route.method != "" && route.method != strings.ToUpper(req.Method) {
			continue
		}
		if route.regex != nil && route.regex.MatchString(req.Path) {
			utils.Debug("Routing to regex route: %s", route.pattern)
			return route.handler(req)
		} else if strings.HasPrefix(route.pattern, ":") || strings.Contains(route.pattern, ":") {
			params := extractParams(route.pattern, req.Path)
			if params != nil {
				req.Params = params
				utils.Debug("Routing to parameterized route: %s", route.pattern)
				return route.handler(req)
			}
		} else if route.pattern == req.Path {
			utils.Debug("Routing to exact match: %s", route.pattern)
			return route.handler(req)
		} else if route.isPrefix && strings.HasPrefix(req.Path, route.pattern) {
			utils.Debug("Routing to prefix route: %s", route.pattern)
			return route.handler(req)
		}
	}
	for _, group := range r.groups {
		for _, route := range group.routes {
			if route.method != "" && route.method != strings.ToUpper(req.Method) {
				continue
			}
			if route.pattern == req.Path {
				return route.handler(req)
			}
		}
	}
	utils.Warn("Route not found for method: %s %s", req.Method, req.Path)
	return NotFoundResponse()
}

func (r *Router) HandlePrefix(prefix, method string, handler HandlerFunc) {
	method = strings.ToUpper(method)
	route := &Route{
		pattern:  prefix,
		method:   method,
		handler:  handler,
		isPrefix: true,
	}
	r.routes = append(r.routes, route)
	utils.Debug("Registered prefix route: %s %s", method, prefix)
}

func extractParams(pattern, path string) map[string]string {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	if len(patternParts) != len(pathParts) {
		return nil
	}

	params := make(map[string]string)
	for i := range patternParts {
		if strings.HasPrefix(patternParts[i], ":") {
			params[patternParts[i][1:]] = pathParts[i]
		} else if patternParts[i] != pathParts[i] {
			return nil
		}
	}
	return params
}

func GetAllowedMethods(methods map[string]HandlerFunc) string {
	var allowed []string
	for m := range methods {
		allowed = append(allowed, m)
	}
	return strings.Join(allowed, ", ")
}

// MethodNotAllowedResponse generates a 405 Method Not Allowed response.
func MethodNotAllowedResponse(allow string) Response {
	return Response{
		Version: "HTTP/1.1",
		Status:  405,
		Reason:  "Method Not Allowed",
		Headers: map[string]string{
			"Content-Type": "text/plain",
			"Allow":        allow,
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
