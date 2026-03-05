package server

import (
	"net/http"

	"github.com/karma-234/vault-mind-mcp/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ServerConfig struct {
	db storage.Storage
}

func NewServer(config ServerConfig) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "vault-mind",
		Version: "0.1.0",
	}, nil)
	return server
}

func NewHTTPHandler(server *mcp.Server) http.Handler {
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
}
