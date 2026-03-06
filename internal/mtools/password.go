package mtools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/jsonschema-go/jsonschema"
	password "github.com/karma-234/vault-mind-mcp/internal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type PasswordParams struct {
	Length         int  `json:"length,omitempty"`
	IncludeSymbols bool `json:"include-symbols,omitempty"`
}
type GenerateMultiplePasswordsParams struct {
	Count          int  `json:"count" jsonschema:"Number of passwords to generate (1-100)"`
	Length         int  `json:"length,omitempty" jsonschema:"Password length (default: 20)"`
	IncludeSymbols bool `json:"include-symbols,omitempty" jsonschema:"Include special characters (default: true)"`
}

func RegisterPasswordTools(s *mcp.Server) {
	schema, err := jsonschema.For[PasswordParams](nil)
	if err != nil {
		log.Fatalf("failed to derive input schema: %v", err)
	}
	s.AddTool(&mcp.Tool{
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

	multiPassSchema, err := jsonschema.For[GenerateMultiplePasswordsParams](nil)
	if err != nil {
		log.Fatalf("failed to derive input schema for generate-multiple-passwords: %v", err)
	}

	s.AddTool(&mcp.Tool{
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
}
