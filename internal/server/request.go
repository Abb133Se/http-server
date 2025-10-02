package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
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
		contentLength, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length: %w", err)
		}
		body := make([]byte, contentLength)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		req.Body = body
	}

	return req, nil
}
