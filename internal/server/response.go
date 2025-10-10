package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/Abb133Se/httpServer/internal/utils"
)

// Response represents an HTTP response message.
//
// It models the key components of an HTTP/1.1 response, including:
//
//   - Version: HTTP version (usually "HTTP/1.1").
//   - Status:  Numeric status code (e.g., 200, 404, 500).
//   - Reason:  Short textual reason phrase associated with the status code.
//   - Headers: Response headers as a key-value map.
//   - Body:    The response body content as a string.
type Response struct {
	Version    string
	Status     int
	Reason     string
	Headers    map[string]string
	Body       []byte
	StreamFunc func(io.Writer) error
}

type ChunkedWriter struct {
	w *bufio.Writer
}

// BuildResponse constructs a raw HTTP response string from the provided parameters.
//
// This function generates a correctly formatted HTTP/1.1 response message
// including the status line, headers, and body. It automatically ensures
// that the `Content-Length` and `Content-Type` headers are set if they are
// not provided by the caller.
//
// Parameters:
//   - status:  HTTP status code (e.g., 200, 404, 500).
//   - reason:  Short description associated with the status code (e.g., "OK", "Not Found").
//   - headers: Optional headers as a map[string]string. If nil, a new map will be created.
//   - body:    The response payload. Its length determines the Content-Length header.
//
// Returns:
//   - string: A raw HTTP response ready to be sent over a TCP connection.
//
// Behavior:
//   - Automatically sets "Content-Length" based on body size.
//   - Defaults "Content-Type" to "text/plain" if none is specified.
//   - Constructs the response in the correct HTTP/1.1 format.
//
// Example:
//
//	raw := server.BuildResponse(200, "OK", map[string]string{
//	    "Content-Type": "text/html",
//	}, "<h1>Hello</h1>")
//
//	fmt.Println(raw)
//	// Output:
//	// HTTP/1.1 200 OK\r\n
//	// Content-Length: 13\r\n
//	// Content-Type: text/html\r\n
//	// \r\n
//	// <h1>Hello</h1>
func BuildResponse(status int, reason string, headers map[string]string, body []byte) []byte {
	if headers == nil {
		headers = make(map[string]string)
	}

	// If not set by handler, set default headers
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "text/plain"
	}
	if _, ok := headers["Content-Length"]; !ok {
		headers["Content-Length"] = strconv.Itoa(len(body))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s %d %s%s", HTTPVersion, status, reason, CRLF))
	for k, v := range headers {
		sb.WriteString(fmt.Sprintf("%s: %s%s", k, v, CRLF))
	}
	sb.WriteString(CRLF)

	// Combine headers + body
	response := append([]byte(sb.String()), body...)
	utils.Debug("Built response: %d %s, Content-Length: %d", status, reason, len(body))
	return response
}

// SendResponse writes an HTTP response to a TCP connection.
//
// This function takes a Response struct, builds the corresponding raw HTTP
// message, and writes it to the provided network connection. It is typically
// used inside a connection handler after processing a request.
//
// Parameters:
//   - conn: The network connection representing the client.
//   - res:  The HTTP response to send.
//
// Returns:
//   - error: Any error encountered while writing to the connection.
//
// Behavior:
//   - Calls BuildResponse to format the response.
//   - Sends the response over the TCP connection.
//
// Example:
//
//	res := server.Response{
//	    Version: server.HTTPVersion,
//	    Status:  200,
//	    Reason:  "OK",
//	    Headers: map[string]string{"Content-Type": "text/plain"},
//	    Body:    "Hello, world!",
//	}
//
//	if err := server.SendResponse(conn, res); err != nil {
//	    log.Printf("failed to send response: %v", err)
//	}
func SendResponse(conn net.Conn, res Response) error {
	writer := bufio.NewWriter(conn)

	fmt.Fprintf(writer, "%s %d %s%s", res.Version, res.Status, res.Reason, CRLF)

	if res.StreamFunc != nil {
		res.Headers["Transfer-Encoding"] = "chunked"
	}

	for k, v := range res.Headers {
		fmt.Fprintf(writer, "%s: %s%s", k, v, CRLF)
	}
	fmt.Fprintf(writer, "%s", CRLF)
	writer.Flush()

	if res.StreamFunc != nil {
		chunkedWriter := NewChunkedWriter(writer)
		if err := res.StreamFunc(chunkedWriter); err != nil {
			utils.Error("Stram error: %v", err)
			return err
		}
		chunkedWriter.Close()
		return writer.Flush()
	}

	if len(res.Body) > 0 {
		_, err := writer.Write(res.Body)
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

func NewChunkedWriter(w *bufio.Writer) *ChunkedWriter {
	return &ChunkedWriter{w: w}
}

func (cw *ChunkedWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	fmt.Fprintf(cw.w, "%x%s", len(p), CRLF)
	cw.w.Write(p)
	cw.w.WriteString(CRLF)
	return len(p), nil
}

func (cw *ChunkedWriter) Close() error {
	// Write zero-length chunk to signal end
	_, err := cw.w.WriteString("0\r\n\r\n")
	return err
}
