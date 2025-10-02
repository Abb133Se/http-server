package server

import (
	"fmt"
	"mime"
	"net/http"
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
// Supported methods:
//   - GET: Returns the contents of the requested file from the
//     "public" directory with an appropriate Content-Type.
//   - POST: Creates or overwrites a file in the "public" directory
//     with the request body.
//
// Errors:
//   - 400 if no file is specified.
//   - 404 if the file does not exist (GET).
//   - 500 if the server cannot resolve the working directory or
//     read/write the file.
//   - 405 if a method other than GET or POST is used.
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

	cwd, err := os.Getwd()
	if err != nil {
		return Response{
			Version: "HTTP/1.1",
			Status:  500,
			Reason:  "Internal Server Error",
			Headers: map[string]string{"Content-Type": "text/plain"},
			Body:    []byte("Cannot get working directory"),
		}
	}

	filePath := filepath.Join(cwd, "public", parts[1])

	if req.Method == "POST" {
		err := os.WriteFile(filePath, req.Body, 0644)
		if err != nil {
			return Response{
				Version: HTTPVersion,
				Status:  500,
				Reason:  "Internal Server Error",
				Headers: map[string]string{"Content-Type": "text/plain"},
				Body:    []byte(fmt.Sprintf("Failed to write file: %s", err.Error())),
			}
		}
		return Response{
			Version: HTTPVersion,
			Status:  http.StatusCreated,
			Reason:  "Created",
			Headers: map[string]string{"Content-Type": "text/plain"},
			Body:    []byte("File created successfully"),
		}
	}

	if req.Method == "GET" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println(err.Error())
			return Response{
				Version: "HTTP/1.1",
				Status:  http.StatusNotFound,
				Reason:  "Not Found",
				Headers: map[string]string{"Content-Type": "text/plain"},
				Body:    []byte(fmt.Sprintf("File not found: %s", parts[1])),
			}
		}

		// Detect MIME type
		ext := filepath.Ext(filePath)
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		return Response{
			Version: "HTTP/1.1",
			Status:  http.StatusOK,
			Reason:  "OK",
			Headers: map[string]string{
				"Content-Type":   mimeType,
				"Content-Length": strconv.Itoa(len(data)),
			},
			Body: data,
		}
	}

	return Response{
		Version: HTTPVersion,
		Status:  http.StatusMethodNotAllowed,
		Reason:  "Method Not Allowed",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    []byte("Only GET and POST methods are supported on /files"),
	}
}
