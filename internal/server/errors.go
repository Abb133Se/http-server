package server

import "github.com/Abb133Se/httpServer/internal/utils"

// Standard error responses
func BadRequestResponse() Response {
	return Response{
		Version: HTTPVersion,
		Status:  400,
		Reason:  "Bad Request",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    []byte("400 Bad Request"),
	}
}

func InternalServerErrorResponse() Response {
	return Response{
		Version: HTTPVersion,
		Status:  500,
		Reason:  "Internal Server Error",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    []byte("500 Internal Server Error"),
	}
}

func NotFoundResponse() Response {
	return Response{
		Version: HTTPVersion,
		Status:  404,
		Reason:  "Not Found",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    []byte("404 Not Found"),
	}
}

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
