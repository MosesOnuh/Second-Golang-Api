package main

import (
	"os"
	"log"

	"github.com/MosesOnuh/todoTask-Api/server"
	
)

func main() {
 port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	err := server.Run(port)
	if err != nil {
		log.Fatal("Could not start server")
	}
}
	
