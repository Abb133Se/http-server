package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

// StartServer initializes and runs the HTTP server on the specified port.
//
// It binds a TCP listener to the given port (e.g., ":8080"), creates a Router,
// registers default routes, and begins accepting client connections. Each
// connection is handled concurrently in its own goroutine.
//
// Registered routes:
//   - "/" → handleRoot
//   - "/echo/{message}" → handleEcho
//   - "/user-agent" → handleUserAgent
//   - "/files/{filename}" → handleFiles (GET, POST)
//
// Parameters:
//   - port: The address and port to bind the server on (e.g., ":8080" or "127.0.0.1:9090").
//
// Returns:
//   - error: Only returned if the TCP listener fails to start.
//     Otherwise, this function typically blocks indefinitely until terminated.
//
// Behavior:
//   - Logs a startup message when the server begins listening.
//   - Spawns a new goroutine for each accepted client connection.
//   - Continues running until the process is stopped.
//
// Example:
//
//	if err := server.StartServer(":8080"); err != nil {
//	    log.Fatalf("Server failed: %v", err)
//	}
func StartServer(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("Failed to start server in port %v\n %v", port, err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Server started on %s\n", port)

	router := NewRouter()
	router.Handle("/", "GET", handleRoot)
	router.HandlePrefix("/echo/", "GET", handleEcho)
	router.Handle("/user-agent", "GET", handleUserAgent)

	router.HandlePrefix("/files/", "GET", handleFiles)
	router.HandlePrefix("/files/", "POST", handleFiles)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connections: %v\n", err)
			continue
		}
		go handleConnection(conn, router)
	}
}

// handleConnection manages the full lifecycle of a single client connection.
//
// This function runs in its own goroutine for each accepted TCP connection.
// It supports persistent (keep-alive) connections, allowing multiple HTTP
// requests to be served sequentially over the same connection.
//
// Flow:
//  1. Defers closure of the client connection to ensure cleanup.
//  2. Sets a read deadline (5 seconds) to prevent hanging clients.
//  3. Parses the incoming HTTP request using ParseRequest.
//  4. Routes the request through the provided Router to obtain a response.
//  5. Adds a "Connection" header based on the client's request
//     (supports "keep-alive" and "close").
//  6. Sends the HTTP response using SendResponse.
//  7. Continues serving new requests if "Connection: keep-alive" is set.
//  8. Terminates when "Connection: close" is requested or a read/send error occurs.
//
// Parameters:
//   - conn:   The TCP connection representing the active client session.
//   - router: The Router instance responsible for dispatching the request.
//
// Behavior:
//   - On read timeout: Closes the connection after 5 seconds of inactivity.
//   - On parse failure: Logs the error and terminates gracefully.
//   - On client disconnect (EOF): Returns silently.
//   - On keep-alive: Reuses the same connection for multiple requests.
//   - Always closes the connection at the end of execution.
//
// Example:
//
//	// Inside StartServer accept loop
//	go handleConnection(conn, router)
func handleConnection(conn net.Conn, router *Router) {
	defer conn.Close()

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		req, err := ParseRequest(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			fmt.Printf("Failed to parse request: %v\n", req)
			return
		}

		resp := router.Route(req)

		connectionHeader := strings.ToLower(req.Headers["connection"])
		if connectionHeader == "keep-alive" {
			resp.Headers["Connection"] = "keep-alive"
		} else {
			resp.Headers["Connection"] = "close"
		}

		if err := SendResponse(conn, resp); err != nil {
			fmt.Printf("failed to send response: %v\n", err)
			return
		}

		if connectionHeader == "close" {
			return
		}
	}

}
