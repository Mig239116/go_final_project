package main

import (
	"bufio"
	"go_final_project/pkg/server"
	"go_final_project/pkg/db"
	"log"
	"os"
	"strings"
)

// loadEnv загружает переменные из файла .env
func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], strings.Trim(parts[1], `"`))
		}
	}
}

func main() {
	loadEnv()
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "../scheduler.db"
	}
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Failed to initialize db: %v", err)
	}
	defer db.Close()
	log.Printf("Database initialized successfully: %s", dbFile)
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7450"
	}
	if err := server.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}