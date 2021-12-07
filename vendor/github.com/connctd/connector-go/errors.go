package connector

import (
	"encoding/json"
	"net/http"
)

// TODO: should we use api-go/errors.go instead? Also used in the signature validation.
var (
	ErrorBadContentType        = Error{Code: http.StatusBadRequest, Message: "Expected content type to be application/json"}
	ErrorMissingInstanceID     = Error{Code: http.StatusBadRequest, Message: "Instance ID is missing"}
	ErrorMissingInstallationID = Error{Code: http.StatusBadRequest, Message: "Installation ID is missing"}
	ErrorBadRequestBody        = Error{Code: http.StatusBadRequest, Message: "Empty or malformed request body"}
	ErrorInvalidJsonBody       = Error{Code: http.StatusBadRequest, Message: "Request body does not contain valid json"}
	ErrorInstallationNotFound  = Error{Code: http.StatusNotFound, Message: "Installation not found"}
	ErrorInstanceNotFound      = Error{Code: http.StatusNotFound, Message: "Instance not found"}
	ErrorForbidden             = Error{Code: http.StatusForbidden, Message: "Insufficient rights"}
	ErrorUnauthorized          = Error{Code: http.StatusUnauthorized, Message: "Not authorized"}
	ErrorInternal              = Error{Code: http.StatusInternalServerError, Message: "Internal server error"}
)

type Error struct {
	Code    int
	Message string
}

func (e *Error) WriteBody(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	b, err := json.Marshal(e)
	if err != nil {
		w.Write([]byte("{\"err\":\"Failed to marshal error\"}"))
	} else {
		w.Write(b)
	}
}

func (e Error) Error() string {
	return e.Message
}
