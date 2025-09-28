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

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connections: %v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}

// handleConnection manages communication with a single TCP client.
//
// This function runs in its own goroutine per connection. It attempts
// to parse an HTTP request and logs the parsed information to stdout.
//
// Parameters:
//   - conn: The net.Conn object representing the client connection.
//
// Behavior:
//   - Parses the request using ParseRequest.
//   - Logs method, path, version, and headers.
//   - Ensures the connection is closed after use.
//
// Example:
//
//	// Inside StartServer
//	go handleConnection(conn)
func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := ParseRequest(conn)
	if err != nil {
		fmt.Printf("Failed to Parse request: %v", err)
		return
	}

	fmt.Printf("Parsed Request: Method=%s, Path=%s, Version=%s\n", req.Method, req.Path, req.Version)
	for k, v := range req.Headers {
		fmt.Printf("Header: %s=%s\n", k, v)
	}
}
