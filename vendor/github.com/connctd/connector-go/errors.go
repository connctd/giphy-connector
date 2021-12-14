package connector

import (
	"net/http"

	"github.com/connctd/api-go"
)

// Errors used in the service and ConnectorHandler
// The ConnectorHandler expects errors of the type api.Error and will set the status code accordingly.
// Developers can define new errors using api.NewError but this should not be necessary for the connector protocol.
var (
	ErrorBadContentType        = api.NewError("BAD_CONTENT_TYPE", "Expected content type to be application/json", http.StatusBadRequest)
	ErrorMissingInstanceID     = api.NewError("MISSING_INSTANCE_ID", "Instance ID is missing", http.StatusBadRequest)
	ErrorMissingInstallationID = api.NewError("MISSING_INSTALLATION_ID", "Installation ID is missing", http.StatusBadRequest)
	ErrorBadRequestBody        = api.NewError("BAD_REQUEST_BODY", "Empty or malformed request body", http.StatusBadRequest)
	ErrorInvalidJsonBody       = api.NewError("INVALID_JSON_BODY", "Request body does not contain valid json", http.StatusBadRequest)
	ErrorInstallationNotFound  = api.NewError("INSTALLATION_NOT_FOUND", "Installation not found", http.StatusNotFound)
	ErrorInstanceNotFound      = api.NewError("INSTANCE_NOT_FOUND", "Instance not found", http.StatusNotFound)
	ErrorForbidden             = api.NewError("FORBIDDEN", "Insufficient rights", http.StatusForbidden)
	ErrorUnauthorized          = api.NewError("NOT_AUTHORIZED", "Not authorized", http.StatusUnauthorized)
	ErrorInternal              = api.NewError("INTERNAL_SERVER_ERROR", "Internal server error", http.StatusInternalServerError)
)
