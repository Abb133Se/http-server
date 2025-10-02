package server

import (
	"fmt"
	"net"
	"os"
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
		fmt.Println("Failed to start server in port " + port)
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

// handleConnection manages the lifecycle of a single client connection.
//
// Flow:
//  1. Defers closure of the client connection.
//  2. Parses the HTTP request from the connection.
//  3. Passes the request to the router to determine the correct handler.
//  4. Sends back the handler's response.
//  5. Logs errors if parsing or sending fails.
//
// Parameters:
//   - conn:   The network connection representing the client session.
//   - router: The Router instance used to dispatch the request.
//
// Behavior:
//   - On parse failure: Logs the error and terminates gracefully.
//   - On success: Routes the request and sends the corresponding response.
//   - Always closes the connection when finished.
//
// Example:
//
//	// Inside StartServer accept loop
//	go handleConnection(conn, router)
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
