package server

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"

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

func NewHttpServer(handler http.Handler) *http.Server {

	cert, err := os.ReadFile("ca.perm")
	if err != nil {
		log.Fatalf("failed to read certificate: %v", err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(cert)
	return &http.Server{
		Addr:    ":8080",
		Handler: handler,
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			MinVersion: tls.VersionTLS13,
			ClientCAs:  caPool,
		},
	}
}

func NewHTTPHandler(server *mcp.Server) http.Handler {
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
}
