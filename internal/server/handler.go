package server

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Abb133Se/httpServer/internal/utils"
)

// handleRoot handles requests to "/".
//
// Supports GET, HEAD, OPTIONS. Returns a plain text welcome message.
// HEAD requests return headers only. OPTIONS responds with allowed methods.
func handleRoot(req *Request) Response {
	utils.Info("Handling root request: %s %s", req.Method, req.Path)

	switch req.Method {
	case "GET", "HEAD":
		body := []byte("Welcome to my HTTP server")
		headers := map[string]string{"Content-Type": "text/plain"}
		if req.Method == "HEAD" {
			body = nil
		}
		return Response{
			Version: HTTPVersion,
			Status:  200,
			Reason:  "OK",
			Headers: headers,
			Body:    body,
		}
	case "OPTIONS":
		return OptionsResponse("GET, HEAD, OPTIONS")
	default:
		utils.Warn("Unsupported method on root: %s %s", req.Method, req.Path)
		return MethodNotAllowedResponse("GET, HEAD, OPTIONS")
	}
}

// handleEcho handles requests to "/echo/{message}".
//
// Supports GET, HEAD, OPTIONS. Returns the path message as plain text.
// If no message is provided, the response body is empty.
func handleEcho(req *Request) Response {
	parts := strings.SplitN(req.Path, "/echo/", 2)
	message := ""
	if len(parts) == 2 {
		message = parts[1]
	}

	switch req.Method {
	case "GET", "HEAD":
		utils.Info("Echo request: %s %s -> %s", req.Method, req.Path, message)
		body := []byte(message)
		if req.Method == "HEAD" {
			body = nil
		}
		return Response{
			Version: HTTPVersion,
			Status:  200,
			Reason:  "OK",
			Headers: map[string]string{"Content-Type": "text/plain"},
			Body:    body,
		}
	case "OPTIONS":
		return OptionsResponse("GET, HEAD, OPTIONS")
	default:
		utils.Warn("Unsupported methods on echo: %s %s", req.Method, req.Path)
		return MethodNotAllowedResponse("GET, HEAD, OPTIONS")
	}
}

// handleUserAgent handles requests to "/user-agent".
//
// Supports GET, HEAD, OPTIONS. Returns the client's "User-Agent" header
// as plain text. HEAD requests return headers only.
func handleUserAgent(req *Request) Response {
	switch req.Method {
	case "GET", "HEAD":
		ua := req.Headers["user-agent"]
		utils.Info("User-Agent request: %s %s -> %s", req.Method, req.Path, ua)
		body := []byte(ua)
		if req.Method == "HEAD" {
			body = nil
		}
		return Response{
			Version: HTTPVersion,
			Status:  200,
			Reason:  "OK",
			Headers: map[string]string{"Content-Type": "text/plain"},
			Body:    body,
		}
	case "OPTIONS":
		return OptionsResponse("GET, HEAD, OPTIONS")
	default:
		utils.Warn("Unsupported method on user-agent: %s %s", req.Method, req.Path)
		return MethodNotAllowedResponse("GET, HEAD, OPTIONS")
	}
}

// handleFiles handles requests to "/files/{filename}".
//
// Supported Methods:
//   - GET: Returns file content from the "public" directory.
//   - POST/PUT: Creates or overwrites a file with the request body.
//   - DELETE: Deletes the specified file.
//   - HEAD: Returns headers only.
//   - OPTIONS: Returns allowed methods.
//
// Error Handling:
//   - 400 Bad Request: No filename specified.
//   - 404 Not Found: File does not exist (GET/DELETE).
//   - 500 Internal Server Error: Failed to read/write the file.
//   - 405 Method Not Allowed: Unsupported HTTP method.
//
// Returns:
//
//	Response struct with status, headers, and body.
func handleFiles(req *Request) Response {
	parts := strings.SplitN(req.Path, "/files/", 2)
	if len(parts) < 2 || parts[1] == "" {
		utils.Warn("File request with no filename: %s %s", req.Method, req.Path)
		return Response{
			Version: HTTPVersion,
			Status:  400,
			Reason:  "Bad Request",
			Headers: map[string]string{"Content-Type": "text/plain"},
			Body:    []byte("No file specified"),
		}
	}

	filePath := filepath.Join(getPublicDir(), parts[1])

	switch req.Method {
	case "GET", "HEAD":
		data, err := os.ReadFile(filePath)
		if err != nil {
			utils.Warn("File not found: %s", filePath)
			return NotFoundResponse()
		}
		mimeType := mime.TypeByExtension(filepath.Ext(filePath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		utils.Info("Serving file: %s (%s)", filePath, mimeType)

		body := data
		if req.Method == "HEAD" {
			body = nil
		}
		return Response{
			Version: "HTTP/1.1",
			Status:  200,
			Reason:  "OK",
			Headers: map[string]string{
				"Content-Type":   mimeType,
				"Content-Length": strconv.Itoa(len(data)),
			},
			Body: body,
		}

	case "POST", "PUT":
		if err := os.WriteFile(filePath, req.Body, 0644); err != nil {
			utils.Error("Failed to write file: %s, error: %v", filePath, err)
			return Response{
				Version: "HTTP/1.1",
				Status:  500,
				Reason:  "Internal Server Error",
				Headers: map[string]string{"Content-Type": "text/plain"},
				Body:    []byte("Failed to write file"),
			}
		}
		status := 201
		reason := "Created"
		if req.Method == "PUT" {
			status = 200
			reason = "OK"
		}
		utils.Info("File %s successfully written", filePath)
		return Response{
			Version: "HTTP/1.1",
			Status:  status,
			Reason:  reason,
			Headers: map[string]string{"Content-Type": "text/plain"},
			Body:    []byte("File written successfully"),
		}

	case "DELETE":
		if err := os.Remove(filePath); err != nil {
			utils.Error("Failed to delete file: %s, error: %v", filePath, err)
			return NotFoundResponse()
		}
		utils.Info("Deleted file: %s", filePath)
		return Response{
			Version: "HTTP/1.1",
			Status:  204,
			Reason:  "No Content",
			Headers: map[string]string{},
			Body:    nil,
		}

	case "OPTIONS":
		return OptionsResponse("GET, HEAD, OPTIONS")

	default:
		utils.Warn("Unsupported method on file: %s %s", req.Method, req.Path)
		return MethodNotAllowedResponse("GET, HEAD, POST, PUT, DELETE, OPTIONS")
	}
}

func handleUserByID(req *Request) Response {
	utils.Info("Regex route matched: %s", req.Path)
	return Response{
		Version: HTTPVersion,
		Status:  200,
		Reason:  "OK",
		Headers: map[string]string{"Content-Type": "text/plain"},
		Body:    []byte(fmt.Sprintf("Matched user path: %s", req.Path)),
	}
}

func handleStream(req *Request) Response {
	utils.Info("Starting streaming response")

	return Response{
		Version: "HTTP/1.1",
		Status:  200,
		Reason:  "OK",
		Headers: map[string]string{"Content-Type": "text/plain"},
		StreamFunc: func(w io.Writer) error {
			for i := 1; i <= 10; i++ {
				fmt.Fprintf(w, "Chunk %d\n", i)
				time.Sleep(1 * time.Second)
			}
			return nil
		},
	}
}

func getPublicDir() string {
	cwd, _ := os.Getwd()
	publicDir := filepath.Join(cwd, "public")
	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		parent := filepath.Dir(cwd)
		publicDir = filepath.Join(parent, "public")
	}
	return publicDir
}
