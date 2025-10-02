package server

import (
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// handleRoot handles requests to the root ("/") path.
//
// It returns a plain text response with status 200 OK and a
// welcome message in the body.
func handleRoot(req *Request) Response {
	return Response{
		Version: HTTPVersion,
		Status:  200,
		Reason:  "OK",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    []byte("Welcome to my HTTP server"),
	}
}

// handleEcho handles requests to "/echo/{message}".
//
// It extracts the message part of the path and returns it as the
// plain text response body. If no message is provided, the body
// is empty.
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
		Body:    []byte(message),
	}
}

// handleUserAgent handles requests to "/user-agent".
//
// It returns the value of the "User-Agent" request header as plain
// text. Header keys are stored in lowercase internally, so the
// correct key is "user-agent".

func handleUserAgent(req *Request) Response {
	ua := req.Headers["user-agent"]
	return Response{
		Version: HTTPVersion,
		Status:  200,
		Reason:  "OK",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    []byte(ua),
	}
}

// handleFiles handles requests to "/files/{filename}".
//
// Supported Methods:
//   - GET: Reads and returns the requested file from the "public" directory
//     with an appropriate Content-Type.
//   - POST: Creates or overwrites a file in the "public" directory with
//     the request body.
//
// Error Handling:
//   - 400 Bad Request: No file specified.
//   - 404 Not Found: File does not exist (GET).
//   - 500 Internal Server Error: Failed to read/write file.
//   - 405 Method Not Allowed: Any method other than GET or POST.
//
// Returns:
//   - Response: An HTTP response with the appropriate status, headers, and body.
func handleFiles(req *Request) Response {
	parts := strings.SplitN(req.Path, "/files/", 2)
	if len(parts) < 2 || parts[1] == "" {
		return Response{
			Version: "HTTP/1.1",
			Status:  400,
			Reason:  "Bad Request",
			Headers: map[string]string{"Content-Type": "text/plain"},
			Body:    []byte("No file specified"),
		}
	}

	cwd, _ := os.Getwd()
	filePath := filepath.Join(cwd, "public", parts[1])

	switch req.Method {
	case "GET":
		data, err := os.ReadFile(filePath)
		if err != nil {
			return NotFoundResponse()
		}
		ext := filepath.Ext(filePath)
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		return Response{
			Version: "HTTP/1.1",
			Status:  200,
			Reason:  "OK",
			Headers: map[string]string{
				"Content-Type":   mimeType,
				"Content-Length": strconv.Itoa(len(data)),
			},
			Body: data,
		}
	case "POST":
		if err := os.WriteFile(filePath, req.Body, 0644); err != nil {
			return Response{
				Version: "HTTP/1.1",
				Status:  500,
				Reason:  "Internal Server Error",
				Headers: map[string]string{"Content-Type": "text/plain"},
				Body:    []byte("Failed to write file"),
			}
		}
		return Response{
			Version: "HTTP/1.1",
			Status:  201,
			Reason:  "Created",
			Headers: map[string]string{"Content-Type": "text/plain"},
			Body:    []byte("File created successfully"),
		}
	default:
		return MethodNotAllowedResponse()
	}
}
