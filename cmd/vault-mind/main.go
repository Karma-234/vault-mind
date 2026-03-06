package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/karma-234/vault-mind-mcp/internal/crypto"
	"github.com/karma-234/vault-mind-mcp/internal/server"
	"github.com/karma-234/vault-mind-mcp/internal/storage"
)

func main() {

	masterPass := os.Getenv("VAULTMIND_PASSPHRASE")
	if masterPass == "" {
		log.Fatal("VAULTMIND_PASSPHRASE environment variable is required. " +
			"Set it before starting the server, e.g.:\n" +
			"  export VAULTMIND_PASSPHRASE='your-strong-passphrase-here'\n" +
			"Then restart your MCP client.")
	}
	key, err := crypto.DeriveKeyFromPassphrase(masterPass)
	if err != nil {
		log.Fatalf("failed to derive key from passphrase: %v", err)
	}
	defer crypto.ZeroKey(key)

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
	}
	store, err := storage.NewVaultPebbleStorage("./vaultmind.db", key)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	newServer := server.NewServer(&server.ServerConfig{
		Pebble: store,
	})

	log.Println("%--- Start running server ---%")
	handler := server.NewHTTPHandler(newServer)

	httpServer := server.NewHttpServer(handler)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on http://127.0.0.1:8080")
		if err := httpServer.ListenAndServeTLS("server.cert", "server.key"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if store != nil {
		if err := store.Close(); err != nil {
			log.Printf("Error closing storage: %v", err)
		}
	}

	log.Println("Server exited")

}
