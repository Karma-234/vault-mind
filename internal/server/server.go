package server

import (
	"net/http"

	"github.com/karma-234/vault-mind-mcp/internal/mtools"
	"github.com/karma-234/vault-mind-mcp/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ServerConfig struct {
	Pebble storage.Storage
}

func NewServer(config *ServerConfig) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "vault-mind",
		Version: "0.1.0",
	}, nil)
	mtools.RegisterPasswordTools(server)
	mtools.RegisterAddCredentialTools(server, config.Pebble)
	mtools.RegisterCredentialRetrievalTools(server, config.Pebble)
	mtools.RegisterDeleteCredentialTools(server, config.Pebble)
	return server
}

func NewHTTPHandler(server *mcp.Server) http.Handler {
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
}
