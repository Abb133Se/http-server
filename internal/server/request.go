package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

// Request represents a parsed HTTP request.
//
// It holds the essential components of an HTTP/1.1 request:
//   - Method:  The HTTP method (e.g., GET, POST, PUT).
//   - Path:    The requested resource path (e.g., "/", "/user").
//   - Version: The HTTP protocol version (typically "HTTP/1.1").
//   - Headers: A map of header keys (lowercased) to values.
//   - Body:    The request body as a string (if present).
type Request struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    string
}

const (
	// HTTPVersion defines the supported HTTP protocol version.
	HTTPVersion = "HTTP/1.1"

	// CLRF represents the carriage-return line-feed sequence used
	// in HTTP request/response formatting.
	CLRF = "\r\n"
)

// ParseRequest reads and parses an HTTP request from a TCP connection.
//
// The function performs the following steps:
//  1. Reads the request line (method, path, version).
//  2. Reads headers until a blank line is encountered.
//  3. Optionally reads the request body if a Content-Length header is present.
//
// Parameters:
//   - conn: The client connection implementing net.Conn.
//
// Returns:
//   - *Request: A pointer to the populated Request struct.
//   - error:    An error if reading/parsing fails.
//
// Limitations:
//   - Only handles HTTP/1.1 style requests.
//   - Only supports Content-Length bodies (no chunked transfer).
//
// Example:
//
//	req, err := server.ParseRequest(conn)
//	if err != nil {
//	    log.Printf("Failed to parse request: %v", err)
//	} else {
//	    fmt.Println("Method:", req.Method)
//	    fmt.Println("Path:", req.Path)
//	}
func ParseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read request lines: %w", err)
	}

	requestLine = strings.TrimSpace(requestLine)

	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed request line: %s", requestLine)
	}

	req := &Request{
		Method:  parts[0],
		Path:    parts[1],
		Version: parts[2],
		Headers: make(map[string]string),
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			key := strings.TrimSpace(headerParts[0])
			value := strings.TrimSpace(headerParts[1])
			req.Headers[strings.ToLower(key)] = value
		}
	}

	if val, ok := req.Headers["content-length"]; ok {
		var bodyBuilder strings.Builder
		_, err := io.CopyN(&bodyBuilder, reader, int64(len(val)))
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		req.Body = bodyBuilder.String()
	}

	return req, nil
}
