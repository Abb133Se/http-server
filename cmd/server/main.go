package main

import (
	"github.com/Abb133Se/httpServer/internal/config"
	"github.com/Abb133Se/httpServer/internal/server"
	"github.com/Abb133Se/httpServer/internal/utils"
)

func main() {
	config := config.LoadConfig()
	utils.InitLogger(config.LogLevel)

	utils.Info("Server starting")

	if err := server.StartServer(config.Port, config); err != nil {
		utils.Error("Error starting server: %v", err)
	}
}
