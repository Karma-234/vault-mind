package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/term"
)

func main() {
	// if running under MCP stdio transport, stdin may not be a TTY; only
	// prompt when a terminal is available.
	pass, err := promptMasterPassphrase()
	if err != nil {
		log.Printf("warning: could not read master passphrase: %v", err)
	}
	// do something with pass if needed
	_ = pass

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
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
				Content: []mcp.Content{&mcp.TextContent{Text: generatedPassword}},
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
func promptMasterPassphrase() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", nil
	}

	bytePass, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(bytePass), nil
}
