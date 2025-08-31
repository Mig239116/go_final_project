package server

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"go_final_project/pkg/api"
)

// Start запускает сервер
func Start(port string) error {
	api.Init()
	webDir, err := getWebDir()
	if err != nil {
		return fmt.Errorf("failed to get web directory: %v", err)
	}
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)
	addr := ":" + port
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("Serving files from: %s", webDir)
	return http.ListenAndServe(addr, nil)
}

// getWebDir определяет путь к директории с файлами фронтэнда
func getWebDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	webDir := filepath.Join(projectRoot, "web")
	
	return webDir, nil
}