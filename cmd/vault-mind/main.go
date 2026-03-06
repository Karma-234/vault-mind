package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/joho/godotenv"
	password "github.com/karma-234/vault-mind-mcp/internal"
	"github.com/karma-234/vault-mind-mcp/internal/crypto"
	"github.com/karma-234/vault-mind-mcp/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "vault-mind",
		Version: "0.1.0",
	}, nil)

	type PasswordParams struct {
		Length         int  `json:"length,omitempty"`
		IncludeSymbols bool `json:"include-symbols,omitempty"`
	}

	schema, err := jsonschema.For[PasswordParams](nil)
	if err != nil {
		log.Fatalf("failed to derive input schema: %v", err)
	}

	server.AddTool(&mcp.Tool{
		Name:        "generate-password",
		Description: "Generate a cryptographically secure random password. Use this before storing credentials.",
		InputSchema: schema,
	},
		func(ctx context.Context, ctr *mcp.CallToolRequest) (*mcp.CallToolResult, error) {

			var params PasswordParams
			if err := json.Unmarshal(ctr.Params.Arguments, &params); err != nil {
				return nil, err
			}

			if params.Length == 0 {
				params.Length = 20
			}
			if !params.IncludeSymbols {
				params.IncludeSymbols = true
			}
			generatedPassword, err := password.GenerateSecurePassword(params.Length, params.IncludeSymbols)
			if err != nil {
				return nil, err
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: generatedPassword},
					&mcp.TextContent{Text: "Cryptographically random. Copy immediately."}},
			}, nil
		})

	type GenerateMultiplePasswordsParams struct {
		Count          int  `json:"count" jsonschema:"Number of passwords to generate (1-100)"`
		Length         int  `json:"length,omitempty" jsonschema:"Password length (default: 20)"`
		IncludeSymbols bool `json:"include-symbols,omitempty" jsonschema:"Include special characters (default: true)"`
	}

	multiPassSchema, err := jsonschema.For[GenerateMultiplePasswordsParams](nil)
	if err != nil {
		log.Fatalf("failed to derive input schema for generate-multiple-passwords: %v", err)
	}

	server.AddTool(&mcp.Tool{
		Name:        "generate-multiple-passwords",
		Description: "Generate multiple cryptographically secure random passwords at once. Useful for batch credential setup.",
		InputSchema: multiPassSchema,
	},
		func(ctx context.Context, ctr *mcp.CallToolRequest) (*mcp.CallToolResult, error) {

			var params GenerateMultiplePasswordsParams
			if err := json.Unmarshal(ctr.Params.Arguments, &params); err != nil {
				return nil, err
			}

			if params.Count <= 0 || params.Count > 100 {
				params.Count = 10
			}
			if params.Length == 0 {
				params.Length = 20
			}
			if !params.IncludeSymbols {
				params.IncludeSymbols = true
			}

			var passwords []string
			for i := 0; i < params.Count; i++ {
				pwd, err := password.GenerateSecurePassword(params.Length, params.IncludeSymbols)
				if err != nil {
					return nil, err
				}
				passwords = append(passwords, pwd)
			}

			passwordList, _ := json.MarshalIndent(passwords, "", "  ")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(passwordList)},
					&mcp.TextContent{Text: fmt.Sprintf("Generated %d passwords. All cryptographically random.", params.Count)},
				},
			}, nil
		})

	log.Println("%--- Start running server ---%")
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)

	httpServer := &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server listening on http://127.0.0.1:8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
