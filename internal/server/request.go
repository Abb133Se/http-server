package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/Abb133Se/httpServer/internal/utils"
)

// Request represents an HTTP/1.1 request.
//
// It contains the method, path, protocol version, headers, and
// optional body of the request. Header keys are normalized to
// lowercase.
type Request struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
	Body    []byte
}

const (
	// HTTPVersion is the supported HTTP protocol version.
	HTTPVersion = "HTTP/1.1"

	// CLRF is the carriage-return/line-feed sequence used in HTTP.
	CLRF = "\r\n"

	MaxRequestLineLength = 4096     // 4 KB max for request line
	MaxHeaderLineLength  = 8192     // 8 KB max for each header line
	MaxBodySize          = 10 << 20 // 10 MB max body
)

// ParseRequest reads and parses an HTTP/1.1 request from a TCP connection.
//
// It reads the request line, headers, and optionally the body if a
// valid Content-Length header is present. Chunked transfer encoding
// is not supported.
func ParseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			utils.Debug("Client closed connection before sending request")
			return nil, err
		}
		utils.Error("Failed to read request line: %v", err)
		return nil, fmt.Errorf("failed to read request lines: %w", err)
	}

	requestLine = strings.TrimSpace(requestLine)
	if len(requestLine) > MaxRequestLineLength {
		utils.Warn("Request line too long: %d bytes", len(requestLine))
		return nil, fmt.Errorf("request line too long")
	}

	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		utils.Warn("Malformed request line: %s", requestLine)
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
			utils.Error("Failed to read header: %v", err)
			return nil, fmt.Errorf("failed to read header: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		if len(line) > MaxHeaderLineLength {
			utils.Warn("Header line too long: %d bytes", len(line))
			return nil, fmt.Errorf("header line too long")
		}

		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			key := strings.TrimSpace(headerParts[0])
			value := strings.TrimSpace(headerParts[1])
			req.Headers[strings.ToLower(key)] = value
		} else {
			utils.Warn("Skipping malformed header line: %s", line)
		}
	}

	if val, ok := req.Headers["content-length"]; ok {
		contentLength, err := strconv.Atoi(val)
		if err != nil {
			utils.Error("Invalid Content-Length: %v", err)
			return nil, fmt.Errorf("invalid Content-Length: %w", err)
		}

		if contentLength > MaxBodySize {
			utils.Warn("Request body too large: %d bytes", contentLength)
			return nil, fmt.Errorf("request body too large")
		}

		body := make([]byte, contentLength)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			utils.Error("Failed to read request body: %v", err)
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		req.Body = body
	}

	utils.Debug("Parsed request: method=%s, path=%s, headers=%v", req.Method, req.Path, req.Headers)
	return req, nil
}
