package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/Husaria55/stock-exchange/internal/handler"
	"github.com/Husaria55/stock-exchange/internal/repository"
	"github.com/Husaria55/stock-exchange/internal/service"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "stockuser")
	dbPass := getEnv("DB_PASSWORD", "stockpassword")
	dbName := getEnv("DB_NAME", "stockexchange")
	port := getEnv("PORT", "8080")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open db connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping db: %v", err)
	}

	repo := repository.NewPostgresRepo(db)
	svc := service.NewStockService(repo)
	router := handler.NewRouter(svc)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	log.Printf("Server starting on port %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}