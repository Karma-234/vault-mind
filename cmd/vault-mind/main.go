package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"os"

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
		if store == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: Storage not initialized. Set VAULTMIND_DB_PATH to a writable directory."},
				},
				IsError: true,
			}, nil
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
	log.Println("%--- Start running server ---%")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
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
