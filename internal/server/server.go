package server

import (
	"fmt"
	"net"
	"os"
)

// StartServer initializes and runs a TCP server on the specified port.
//
// It listens for incoming client connections and spawns a new goroutine
// for each accepted connection using handleConnection. The server runs
// indefinitely until terminated by the user (e.g., SIGINT) or an error occurs.
//
// Parameters:
//   - port: A string specifying the port/address to bind the server on,
//     e.g. ":8080" or "127.0.0.1:9090".
//
// Returns:
//   - error: If the server fails to bind to the given port. On success,
//     this function does not return under normal execution flow.
//
// Behavior:
//   - Logs a message to stdout when the server starts.
//   - For each accepted connection, starts a goroutine to handle it.
//   - Exits the process immediately if the port binding fails.
//
// Example:
//
//	err := server.StartServer(":8080")
//	if err != nil {
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
