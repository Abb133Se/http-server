package main

import (
	"fmt"
	"log"

	"github.com/Abb133Se/httpServer/internal/config"
	"github.com/Abb133Se/httpServer/internal/server"
)

func main() {
	fmt.Println("server running: ")

	config := config.LoadConfig()

	if err := server.StartServer(config.Port); err != nil {
		log.Fatalf("Error Starting Server: %v", err)
	}
}
