package server

import (
	"fmt"
	"net"
	"os"
)

// StartServer initializes and runs the HTTP server on the specified address.
//
// It binds a TCP listener to the given address (e.g., ":8080"), creates a
// Router, registers the default routes, and begins accepting client
// connections. Each connection is handled in a separate goroutine.
//
// Registered routes:
//   - "/" → handleRoot
//   - "/echo/{message}" → handleEcho
//   - "/user-agent" → handleUserAgent
//
// Parameters:
//   - addr: The address and port to bind the server on, e.g., ":8080" or "127.0.0.1:9090".
//
// Returns:
//   - error: A wrapped error if the listener fails to bind or is closed unexpectedly.
//     On success, this function typically blocks indefinitely.
//
// Behavior:
//   - Logs a startup message when the server begins listening.
//   - Spawns a new goroutine for each client connection.
//   - Continues serving until terminated externally (e.g., SIGINT).
//
// Example:
//
//	if err := server.StartServer(":8080"); err != nil {
//	    log.Fatalf("Server failed: %v", err)
//	}
func StartServer(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Failed to start server in port " + port)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Server started on %s\n", port)

	router := NewRouter()
	router.Handle("/", handleRoot)
	router.HandlePrefix("/echo/", handleEcho)
	router.Handle("/user-agent", handleUserAgent)

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
// This function is executed in its own goroutine for each accepted TCP client.
// It is responsible for parsing the HTTP request and sending back an appropriate
// HTTP response.
//
// Flow:
//  1. Defer closure of the connection to ensure cleanup.
//  2. Parse the incoming HTTP request using ParseRequest.
//  3. Construct an HTTP response with status 200 OK, a plain-text content type,
//     and a body that echoes the requested path.
//  4. Send the response using SendResponse.
//  5. Log any errors encountered along the way.
//
// Parameters:
//   - conn: The network connection representing the client session.
//
// Behavior:
//   - On request parse failure: Logs the error and terminates gracefully.
//   - On successful request: Returns a "Hello! You requested {path}" message.
//   - Always closes the connection at the end of execution.
//
// Example:
//
//	// Inside StartServer accept loop
//	go handleConnection(conn)
//
// Client Request:
//
//	GET /greet HTTP/1.1
//	Host: localhost:8080
//
// Server Response:
//
//	HTTP/1.1 200 OK
//	Content-Length: 28
//	Content-Type: text/plain
//
//	Hello! You requested /greet
func handleConnection(conn net.Conn, router *Router) {
	defer conn.Close()

	req, err := ParseRequest(conn)
	if err != nil {
		fmt.Printf("Failed to Parse request: %v", err)
		return
	}

	resp := router.Route(req)

	if err := SendResponse(conn, resp); err != nil {
		fmt.Printf("failed to send response: %v\n", err)
	}
}
