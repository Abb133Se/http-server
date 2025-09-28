package server

import "strings"

// handleRoot is the default handler for the root ("/") path.
//
// Returns:
//   - A Response with HTTP status 200 OK.
//   - Content-Type set to "text/plain".
//   - Body message welcoming the client to the server.
//
// Example Response:
//
//	HTTP/1.1 200 OK
//	Content-Type: text/plain
//	Content-Length: 27
//
//	Welcome to my HTTP server
func handleRoot(req *Request) Response {
	return Response{
		Version: HTTPVersion,
		Status:  200,
		Reason:  "OK",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    "Welcome to my HTTP server",
	}
}

// handleEcho is the handler for the "/echo/{message}" path.
//
// Behavior:
//   - Extracts the substring after "/echo/" from the request path.
//   - Returns the extracted message as the body of the response.
//   - If no message is provided, the body is an empty string.
//
// Example Request:
//
//	GET /echo/hello HTTP/1.1
//
// Example Response:
//
//	HTTP/1.1 200 OK
//	Content-Type: text/plain
//	Content-Length: 5
//
//	hello
func handleEcho(req *Request) Response {
	parts := strings.SplitN(req.Path, "/echo/", 2)
	message := ""
	if len(parts) == 2 {
		message = parts[1]
	}

	return Response{
		Version: HTTPVersion,
		Status:  200,
		Reason:  "OK",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    message,
	}
}
