package main

import (
	"fmt"
	"log"

	"github.com/Abb133Se/httpServer/internal/server"
)

func main() {
	fmt.Println("server running: ")
	if err := server.StartServer(":4221"); err != nil {
		log.Fatalf("Error Starting Server: %v", err)
	}
}
