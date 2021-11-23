package http

import (
	"context"
	"crypto/ed25519"
	"giphy-connector"
	"net/http"

	"github.com/connctd/connector-go"
	"github.com/gorilla/mux"
)

// MakeHandler returns a new HTTP handler that provides the endpoint required for the connector protocol
func MakeHandler(ctx context.Context, publicKey ed25519.PublicKey, service giphy.Service) http.Handler {
	r := mux.NewRouter()

	setupConnectorProtocolCallbacks(r.PathPrefix("/callbacks").Subrouter(), publicKey, service)
	return r
}

// setupConnectorProtocolCallbacks sets up the routes needed for the connector protocol
// We are using connector.NewSignatureValidationHandler to verify the signature before our handler are called.
// For more information on the signature see https://docs.connctd.io/connector/connector_protocol/#installation-callback
// Note that the specific path is not part of the protocol. Instead the complete URL can be specified during connector publication.
// Feel free to change the path (and the prefix defined in MakeHandler) as you like
// Also note, that the connector-go SDK provides convenience functions that will handle this setup for you
// We implement it manually for demonstration purposes
func setupConnectorProtocolCallbacks(r *mux.Router, publicKey ed25519.PublicKey, service giphy.Service) {
	r.Path("/installations").Methods(http.MethodPost).Handler(
		connector.NewSignatureValidationHandler(
			connector.DefaultValidationPreProcessor(),
			publicKey,
			handleInstallation(service),
		),
	)
	r.Path("/instantiations").Methods(http.MethodPost).Handler(
		connector.NewSignatureValidationHandler(
			connector.DefaultValidationPreProcessor(),
			publicKey,
			handleInstantiation(service),
		),
	)
	r.Path("/actions").Methods(http.MethodPost).Handler(
		connector.NewSignatureValidationHandler(
			connector.DefaultValidationPreProcessor(),
			publicKey,
			handleAction(service),
		),
	)
}
