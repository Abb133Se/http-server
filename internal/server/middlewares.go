package server

import "github.com/Abb133Se/httpServer/internal/utils"

func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(req *Request) Response {
		utils.Info("Middleware: %s %s", req.Method, req.Path)
		resp := next(req)
		utils.Info("Response status: %d %s", resp.Status, resp.Reason)
		return resp
	}
}
