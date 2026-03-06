package mtools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/karma-234/vault-mind-mcp/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetCredentialSchema struct {
	ID string `json:"id" jsonschema:"The unique ID of the credential to retrieve."`
}

func RegisterCredentialRetrievalTools(s *mcp.Server, store storage.Storage) {
	s.AddTool(&mcp.Tool{
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
	getCredSchema, err := jsonschema.For[GetCredentialSchema](nil)
	if err != nil {
		log.Fatalf("failed to derive input schema for get-credential: %v", err)
	}
	s.AddTool(&mcp.Tool{
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
}
