package connector

import (
	"encoding/json"
	"net/http"
)

// Errors used in the service and ConnectorHandler
// The ConnectorHandler expects errors of the type connector.Error and will set the status code accordingly.
// Developers can define new errors using connector.NewError but this should not be necessary for the connector protocol.
var (
	ErrorBadContentType        = NewError("BAD_CONTENT_TYPE", "Expected content type to be application/json", http.StatusBadRequest)
	ErrorMissingInstanceID     = NewError("MISSING_INSTANCE_ID", "Instance ID is missing", http.StatusBadRequest)
	ErrorMissingInstallationID = NewError("MISSING_INSTALLATION_ID", "Installation ID is missing", http.StatusBadRequest)
	ErrorBadRequestBody        = NewError("BAD_REQUEST_BODY", "Empty or malformed request body", http.StatusBadRequest)
	ErrorInvalidJsonBody       = NewError("INVALID_JSON_BODY", "Request body does not contain valid json", http.StatusBadRequest)
	ErrorInstallationNotFound  = NewError("INSTALLATION_NOT_FOUND", "Installation not found", http.StatusNotFound)
	ErrorInstanceNotFound      = NewError("INSTANCE_NOT_FOUND", "Instance not found", http.StatusNotFound)
	ErrorForbidden             = NewError("FORBIDDEN", "Insufficient rights", http.StatusForbidden)
	ErrorUnauthorized          = NewError("NOT_AUTHORIZED", "Not authorized", http.StatusUnauthorized)
	ErrorInternal              = NewError("INTERNAL_SERVER_ERROR", "Internal server error", http.StatusInternalServerError)
)

// NewError constructs an error
func NewError(err string, description string, status int) *Error {
	return &Error{
		APIError:    err,
		Status:      status,
		Description: description,
	}
}

// Error defines an error
type Error struct {
	APIError    string `json:"error"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

// Error returns the errors description
func (e *Error) Error() string {
	return e.Description
}

// Write uses given response writer to write an error
func (e *Error) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.Status)

	b, err := json.Marshal(e)
	if err != nil {
		w.Write([]byte("{\"error\":\"" + err.Error() + "\"}"))
	}

	w.Write(b)
}
