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

type DeleteCredentialParams struct {
	ID string `json:"id" jsonschema:"The unique ID of the credential to delete."`
}

type BulkDeleteCredentialParams struct {
	IDs []string `json:"ids" jsonschema:"A list of credential IDs to delete. Use with extreme caution."`
}

func RegisterDeleteCredentialTools(s *mcp.Server, store storage.Storage) {
	schema, err := jsonschema.For[DeleteCredentialParams](nil)
	if err != nil {
		log.Fatalf("failed to derive input schema: %v", err)
	}
	s.AddTool(&mcp.Tool{
		Name:        "delete-credential",
		Description: "Delete a credential by its ID. Use ListCredentials to get IDs.",
		InputSchema: schema,
	},
		func(ctx context.Context, ctr *mcp.CallToolRequest) (*mcp.CallToolResult, error) {

			var params DeleteCredentialParams
			if err := json.Unmarshal(ctr.Params.Arguments, &params); err != nil {
				return nil, err
			}

			err := store.DeleteCredential(params.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to delete credential: %w", err)
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Credential deleted successfully."},
				},
			}, nil
		})
	bulkDelSchema, err := jsonschema.For[BulkDeleteCredentialParams](nil)
	if err != nil {
		log.Fatalf("failed to derive input schema: %v", err)
	}
	s.AddTool(&mcp.Tool{
		Name:        "delete-multiple-credentials",
		Description: "Delete multiple stored credentials by Id. Use with extreme caution.",
		InputSchema: bulkDelSchema,
	}, func(ctx context.Context, ctr *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var params BulkDeleteCredentialParams
		if err := json.Unmarshal(ctr.Params.Arguments, &params); err != nil {
			return nil, err
		}

		var failedDeletions []string
		for _, id := range params.IDs {
			err := store.DeleteCredential(id)
			if err != nil {
				failedDeletions = append(failedDeletions, id)
				log.Printf("Failed to delete credential with ID %s: %v", id, err)
			}
		}

		if len(failedDeletions) > 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to delete credentials with IDs: %v", failedDeletions)},
				},
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("%d credentials deleted successfully.", len(params.IDs))},
			},
		}, nil
	})

	s.AddTool(&mcp.Tool{
		Name:        "delete-all-credentials",
		Description: "Delete ALL stored credentials. This action cannot be undone. Use with extreme caution.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(ctx context.Context, ctr *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		err := store.DeleteAllCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to delete all credentials: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "All credentials have been deleted. This action cannot be undone."},
			},
		}, nil
	})
}
