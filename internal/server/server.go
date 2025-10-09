package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/Abb133Se/httpServer/internal/config"
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
func StartServer(port string, config *config.Config) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to start server on port %s: %w", port, err)
	}
	defer listener.Close()

	utils.Info("Server started on %s", port)

	router := NewRouter()
	setupRoutes(router)

	for {
		conn, err := listener.Accept()
		if err != nil {
			utils.Warn("Failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn, router, config)
	}
}

func setupRoutes(router *Router) {
	router.Handle("/", "GET", handleRoot)
	router.Handle("/", "HEAD", handleRoot)
	router.Handle("/", "OPTIONS", handleRoot)

	router.Handle("/echo/", "GET", handleEcho)
	router.Handle("/echo/", "HEAD", handleEcho)
	router.Handle("/echo/", "OPTIONS", handleEcho)

	router.Handle("/user-agent", "GET", handleUserAgent)
	router.Handle("/user-agent", "HEAD", handleUserAgent)
	router.Handle("/user-agent", "OPTIONS", handleUserAgent)

	router.HandlePrefix("/files/", "GET", handleFiles)
	router.HandlePrefix("/files/", "HEAD", handleFiles)
	router.HandlePrefix("/files/", "POST", handleFiles)
	router.HandlePrefix("/files/", "PUT", handleFiles)
	router.HandlePrefix("/files/", "DELETE", handleFiles)
	router.HandlePrefix("/files/", "OPTIONS", handleFiles)

	router.HandleRegex(`^/user/\d+$`, handleUserByID)

	utils.Info("All routes registered successfully")
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
func handleConnection(conn net.Conn, router *Router, config *config.Config) {
	defer conn.Close()

	for {
		conn.SetReadDeadline(time.Now().Add(config.ReadTimeout))

		req, err := ParseRequest(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				utils.Debug("Connection closed by client")
				return
			}
			utils.Warn("Malformed or oversized request: %v", err)
			resp := Response{
				Version: HTTPVersion,
				Status:  400,
				Reason:  "Bad Request",
				Headers: map[string]string{"Content-Type": "text/plain"},
				Body:    []byte("400 Bad Request"),
			}

			if sendErr := SendResponse(conn, resp); sendErr != nil {
				utils.Warn("Failed to send 400 response: %v", sendErr)
			}
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
