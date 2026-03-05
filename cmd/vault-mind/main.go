package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/joho/godotenv"
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
			generatedPassword, err := generateSecurePassword(params.Length, params.IncludeSymbols)
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
				pwd, err := generateSecurePassword(params.Length, params.IncludeSymbols)
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

	type AddCredentialParams struct {
		Service string `json:"service" jsonschema:"Service name (e.g., 'GitHub')"`
		Type    string `json:"type" jsonschema:"Credential type (e.g., 'API Key', 'Seed Phrase')"`
		Secret  string `json:"secret" jsonschema:"The sensitive credential value. It will be encrypted before storage."`
		Notes   string `json:"notes,omitempty" jsonschema:"Additional notes (optional)"`
	}
	addCredSchema, err := jsonschema.For[AddCredentialParams](nil)
	server.AddTool(&mcp.Tool{
		Name:        "add-credential",
		Description: "Securely add a credential (e.g., API key, seed phrase). It's encrypted before storage.",
		InputSchema: addCredSchema,
	}, func(ctx context.Context, ctr *mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		var params AddCredentialParams
		if err := json.Unmarshal(ctr.Params.Arguments, &params); err != nil {
			return nil, err
		}
		_, err := store.AddCredential(params.Service, params.Type, string(params.Secret), params.Notes)
		if err != nil {
			return nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Credential added successfully."},
			},
		}, nil
	})

	type BulkCredentialParams struct {
		Entries []storage.AddCredentialInput `json:"entries" jsonschema:"A list of credentials to store"`
	}
	bulkSchema, err := jsonschema.For[BulkCredentialParams](nil)
	if err != nil {
		log.Fatalf("failed to derive input schema for add-bulk-credentials: %v", err)
	}
	server.AddTool(&mcp.Tool{
		Name:        "add-bulk-credentials",
		Description: "Add multiple credentials in a single call. Each entry is encrypted individually. Optimized for batch operations.",
		InputSchema: bulkSchema,
	}, func(ctx context.Context, ctr *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params BulkCredentialParams
		if err := json.Unmarshal(ctr.Params.Arguments, &params); err != nil {
			return nil, err
		}
		if err := store.AddBulkCredentials(params.Entries); err != nil {
			return nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("%d credentials added in bulk.", len(params.Entries))},
			},
		}, nil
	})

	server.AddTool(&mcp.Tool{
		Name:        "list-credentials",
		Description: "List all stored credentials with their service, type, and notes. Does not reveal secrets.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, ctr *mcp.CallToolRequest) (*mcp.CallToolResult, error) {

		creds, err := store.ListCredentials()
		if err != nil {
			return nil, err
		}
		var content []mcp.Content
		for _, cred := range creds {
			content = append(content, &mcp.TextContent{
				Text: fmt.Sprintf("ID: %s | Service: %s | Type: %s | Notes: %s | Created: %s",
					cred.ID, cred.Service, cred.Type, cred.Notes, cred.Created.Format(time.RFC3339)),
			})
		}
		return &mcp.CallToolResult{
			Content: content,
		}, nil
	})
	type GetCredentialSchema struct {
		ID string `json:"id" jsonschema:"The unique ID of the credential to retrieve."`
	}
	getCredSchema, err := jsonschema.For[GetCredentialSchema](nil)
	if err != nil {
		log.Fatalf("failed to derive input schema for get-credential: %v", err)
	}
	server.AddTool(&mcp.Tool{
		Name:        "get-credential",
		Description: "Retrieve the secret for a specific credential by ID. Use with caution to avoid exposing secrets.",
		InputSchema: getCredSchema,
	}, func(ctx context.Context, ctr *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params GetCredentialSchema
		if err := json.Unmarshal(ctr.Params.Arguments, &params); err != nil {
			return nil, err
		}
		cred, plaintext, err := store.GetCredential(params.ID)
		if err != nil {
			return nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Service: %s | Type: %s | Notes: %s | Created: %s",
					cred.Service, cred.Type, cred.Notes, cred.Created.Format(time.RFC3339))},
				&mcp.TextContent{Text: fmt.Sprintf("Secret: %s", plaintext)},
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

func generateSecurePassword(length int, includeSymbols bool) (string, error) {
	letters := os.Getenv("LETTERS")
	if letters == "" {
		letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}

	symbols := os.Getenv("SYMBOLS")
	if symbols == "" {
		symbols = "!@#$%^&*()-_=+[]{}|;:,.<>?/~`"
	}

	charCycle := letters
	if includeSymbols {
		charCycle += symbols
	}

	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charCycle))))
		if err != nil {
			return "", err
		}
		b[i] = charCycle[num.Int64()]
	}
	return string(b), nil
}
