package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/config"
	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/handlers"
	"github.com/kurnosovmak/akurnosov-microservice-auth/internal/storage"
)

func init() {
	config.LoadEnv()
}

func main() {
	// Инициализация базы данных
	if err := storage.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Создание роутера
	r := mux.NewRouter()

	// Middleware
	r.Use(handlers.LoggingMiddleware)
	r.Use(handlers.CORSMiddleware)
	r.Use(handlers.RecoveryMiddleware)

	// Маршруты
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/verify", handlers.VerifyHandler).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")

	// Создание HTTP сервера
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Println("Server started at http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown failed:", err)
	}
	log.Println("Server stopped gracefully")
}
