package main

import (
	"log"
	"net/http"

	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/config"
	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/handlers"
	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/storage"
)

func main() {
	config.LoadEnv()
	if err := storage.InitDB(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/verify", handlers.VerifyHandler)
	http.HandleFunc("/login", handlers.LoginHandler)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
