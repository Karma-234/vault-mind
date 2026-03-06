package mtools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/karma-234/vault-mind-mcp/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AddCredentialParams struct {
	Service string `json:"service" jsonschema:"Service name (e.g., 'GitHub')"`
	Type    string `json:"type" jsonschema:"Credential type (e.g., 'API Key', 'Seed Phrase')"`
	Secret  string `json:"secret" jsonschema:"The sensitive credential value. It will be encrypted before storage."`
	Notes   string `json:"notes,omitempty" jsonschema:"Additional notes (optional)"`
}
type BulkCredentialParams struct {
	Entries []storage.AddCredentialInput `json:"entries" jsonschema:"A list of credentials to store"`
}

func RegisterAddCredentialTools(s *mcp.Server, store storage.Storage) {
	addCredSchema, err := jsonschema.For[AddCredentialParams](nil)

	s.AddTool(&mcp.Tool{
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

	bulkSchema, err := jsonschema.For[BulkCredentialParams](nil)
	if err != nil {
		log.Fatalf("failed to derive input schema for add-bulk-credentials: %v", err)
	}
	s.AddTool(&mcp.Tool{
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

}
