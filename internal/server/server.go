package server

import (
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Abb133Se/httpServer/internal/utils"
)

// StartServer starts a TCP-based HTTP server on the specified port.
//
// It sets up a Router, registers standard routes, and listens for incoming
// client connections. Each connection is handled in its own goroutine, supporting
// persistent connections (keep-alive) when requested.
//
// Supported routes:
//   - "/" → handleRoot
//   - "/echo/{message}" → handleEcho
//   - "/user-agent" → handleUserAgent
//   - "/files/{filename}" → handleFiles (GET, POST, PUT, DELETE, HEAD, OPTIONS)
//
// Parameters:
//   - port: The address and port to bind the server on (e.g., ":8080").
//
// Returns:
//   - error: Only if the TCP listener fails to start. Otherwise, this function
//     blocks indefinitely until externally terminated.
//
// Example:
//
//	if err := server.StartServer(":8080"); err != nil {
//	    log.Fatalf("Server failed: %v", err)
//	}
func StartServer(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		utils.Error("Failed to start server on port %s: %v", port, err)
		os.Exit(1)
	}
	defer listener.Close()

	utils.Info("Server started on %s", port)

	router := NewRouter()
	router.Handle("/", "GET", handleRoot)
	router.Handle("/", "HEAD", handleRoot)
	router.Handle("/", "OPTIONS", handleRoot)

	router.HandlePrefix("/echo/", "GET", handleEcho)
	router.HandlePrefix("/echo/", "HEAD", handleEcho)
	router.HandlePrefix("/echo/", "OPTIONS", handleEcho)

	router.Handle("/user-agent", "GET", handleUserAgent)
	router.HandlePrefix("/echo/", "HEAD", handleUserAgent)
	router.HandlePrefix("/echo/", "OPTIONS", handleUserAgent)

	router.HandlePrefix("/files/", "GET", handleFiles)
	router.HandlePrefix("/files/", "HEAD", handleFiles)
	router.HandlePrefix("/files/", "POST", handleFiles)
	router.HandlePrefix("/files/", "PUT", handleFiles)
	router.HandlePrefix("/files/", "DELET", handleFiles)
	router.HandlePrefix("/files/", "OPTIONS", handleFiles)

	for {
		conn, err := listener.Accept()
		if err != nil {
			utils.Warn("Failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn, router)
	}
}

// handleConnection manages the lifecycle of a single client TCP connection.
//
// It supports persistent connections (HTTP keep-alive) and sequentially
// serves multiple requests on the same connection if requested.
//
// Flow:
//  1. Sets a read deadline of 5 seconds to prevent hanging connections.
//  2. Parses the HTTP request using ParseRequest.
//  3. Routes the request via the provided Router.
//  4. Adds the appropriate "Connection" header based on the request.
//  5. Sends the response and repeats if "Connection: keep-alive".
//  6. Terminates on "Connection: close" or any read/send error.
//
// Parameters:
//   - conn: TCP connection representing the client session.
//   - router: Router instance responsible for dispatching requests.
//
// Behavior:
//   - Closes the connection after inactivity or errors.
//   - Logs requests and responses with status and reason.
//   - Handles EOF gracefully when the client disconnects.
//
// Example:
//
//	go handleConnection(conn, router)
func handleConnection(conn net.Conn, router *Router) {
	defer conn.Close()

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		req, err := ParseRequest(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				utils.Debug("Connection closed by client")
				return
			}
			utils.Warn("Failed to parse request: %v", err)
			return
		}
		utils.Info("Incoming request: %s %s", req.Method, req.Path)

		resp := router.Route(req)

		connectionHeader := strings.ToLower(req.Headers["connection"])
		if connectionHeader == "keep-alive" {
			resp.Headers["Connection"] = "keep-alive"
		} else {
			resp.Headers["Connection"] = "close"
		}

		if err := SendResponse(conn, resp); err != nil {
			utils.Warn("Failed to send response: %v", err)
			return
		}

		utils.Info("Response sent: %s %s -> %d %s", req.Method, req.Path, resp.Status, resp.Reason)

		if connectionHeader == "close" {
			utils.Debug("Closing connection as per header")
			return
		}
	}

}
