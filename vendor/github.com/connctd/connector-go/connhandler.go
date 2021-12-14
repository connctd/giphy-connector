package connector

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/connctd/api-go"
	"github.com/gorilla/mux"
)

// ConnectorHandler implements all endpoints used in the connector protocol and validates all incoming requests with the SignatureValidationHandler.
// Connector developers ususally do not need to modify any of the handlers.
type ConnectorHandler struct {
	router  *mux.Router
	service ConnectorService
}

// ServeHTTP implements the http.Handler interface by delegating to the router
func (c *ConnectorHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	c.router.ServeHTTP(w, request)
}

func baseConnectorHandler(subrouter *mux.Router, service ConnectorService) *ConnectorHandler {
	c := &ConnectorHandler{
		router:  subrouter,
		service: service,
	}

	if c.router == nil {
		c.router = mux.NewRouter()
	}

	return c
}

// NewAutoProxyConnectorHandler returns a connector handler that detects proxies and modifies the validation parameters
// for the signature validation. This should be used by default and should also work without any proxies in place.
// Note that the proxy has to set the correct headers for this to work. See AutoProxyRequestValidationPreProcessor for more information.
func NewConnectorHandler(subrouter *mux.Router, service ConnectorService, publicKey ed25519.PublicKey) *ConnectorHandler {

	c := baseConnectorHandler(subrouter, service)

	c.router.Path("/installations").Methods(http.MethodPost).Handler(NewSignatureValidationHandler(
		AutoProxyRequestValidationPreProcessor(), publicKey, AddInstallation(c.service)))
	c.router.Path("/installations/{id}").Methods(http.MethodDelete).Handler(NewSignatureValidationHandler(
		AutoProxyRequestValidationPreProcessor(), publicKey, RemoveInstallation(c.service)))

	c.router.Path("/instances").Methods(http.MethodPost).Handler(NewSignatureValidationHandler(
		AutoProxyRequestValidationPreProcessor(), publicKey, AddInstance(c.service)))
	c.router.Path("/instances/{id}").Methods(http.MethodDelete).Handler(NewSignatureValidationHandler(
		AutoProxyRequestValidationPreProcessor(), publicKey, RemoveInstance(c.service)))

	c.router.Path("/actions").Methods(http.MethodPost).Handler(NewSignatureValidationHandler(
		AutoProxyRequestValidationPreProcessor(), publicKey, PerformAction(c.service)))

	return c
}

// NewConnectorHandler lets you manually set the host used for the signature validation.
// This can be usefull if you are behind a proxy that doesn't set the correct headers.
// Note that we set the protocol to "https" since this is the only protocol supported by connctd.
func NewProxiedConnectorHandler(subrouter *mux.Router, service ConnectorService, host string, publicKey ed25519.PublicKey) *ConnectorHandler {
	c := baseConnectorHandler(subrouter, service)

	c.router.Path("/installations").Methods(http.MethodPost).Handler(NewSignatureValidationHandler(
		ProxiedRequestValidationPreProcessor("https", host), publicKey, AddInstallation(c.service),
	))
	c.router.Path("/installations/{id}").Methods(http.MethodDelete).Handler(NewSignatureValidationHandler(
		ProxiedRequestValidationPreProcessor("https", host), publicKey, RemoveInstallation(c.service),
	))

	c.router.Path("/instances").Methods(http.MethodPost).Handler(NewSignatureValidationHandler(
		ProxiedRequestValidationPreProcessor("https", host), publicKey, AddInstallation(c.service),
	))
	c.router.Path("/instances/{id}").Methods(http.MethodDelete).Handler(NewSignatureValidationHandler(
		ProxiedRequestValidationPreProcessor("https", host), publicKey, RemoveInstance(c.service),
	))

	c.router.Path("/actions").Methods(http.MethodPost).Handler(NewSignatureValidationHandler(
		ProxiedRequestValidationPreProcessor("https", host), publicKey, PerformAction(c.service),
	))

	return c
}

// AddInstallation is called whenever a connector is installed via the connctd platform.
// It will validate the request and delegate valid requests to the service.
// It expects an error from errors.go.
// The status code will be set to one defined in the error and the InstantiationResponse will be returned to the connctd platform.
func AddInstallation(service ConnectorService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var req InstallationRequest
		if err := decodeJSONBody(w, r, &req); err != nil {
			writeError(w, err)
			return
		}

		response, err := service.AddInstallation(r.Context(), req)
		if err != nil {
			writeStatus(w, err)
			if response != nil {
				b, err := json.Marshal(response)
				if err != nil {
					writeError(w, err)
					return
				}
				w.Write(b)
			}
			return
		}

		if response != nil {
			b, err := json.Marshal(response)
			if err != nil {
				writeError(w, err)
				return
			}
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			w.Write(b)
			return
		}
		// We set the content type to application/json to prevent ngrok from interpreting the response as HTML
		// and serving a landing page instead.
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	})
}

// RemoveInstallation is called whenever an installation is removed by the the connctd platform.
func RemoveInstallation(service ConnectorService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, ok := vars["id"]

		if !ok {
			writeError(w, ErrorMissingInstallationID)
			return
		}

		if err := service.RemoveInstallation(r.Context(), id); err != nil {
			writeError(w, err)
			return
		}

		// We set the content type to application/json to prevent ngrok from interpreting the response as HTML
		// and serving a landing page instead.
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})
}

// AddInstance is called whenever a connector is instantiated via the connctd platform.
// It will validate the request and delegate valid requests to the service.
// It expects an error from errors.go.
// The status code will be set to one defined in the error and the InstantiationResponse will be returned to the connctd platform.
func AddInstance(service ConnectorService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req InstantiationRequest
		if err := decodeJSONBody(w, r, &req); err != nil {
			writeError(w, err)
			return
		}

		response, err := service.AddInstance(r.Context(), req)
		if err != nil {
			writeStatus(w, err)
			if response != nil {
				b, err := json.Marshal(response)
				if err != nil {
					writeError(w, err)
					return
				}
				// Do not set a status code, since we want to keep the status set by the error.
				w.Write(b)
			}
			return
		}

		if response != nil {
			b, err := json.Marshal(response)
			if err != nil {
				writeError(w, err)
				return
			}
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			w.Write(b)
			return
		}

		// We set the content type to application/json to prevent ngrok from interpreting the response as HTML
		// and serving a landing page instead.
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

	})
}

//RemoveInstance is called whenever an instance is removed by the the connctd platform.
func RemoveInstance(service ConnectorService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, ok := vars["id"]

		if !ok {
			writeError(w, ErrorMissingInstanceID)
			return
		}

		if err := service.RemoveInstance(r.Context(), id); err != nil {
			writeError(w, err)
			return
		}

		// We set the content type to application/json to prevent ngrok from interpreting the response as HTML
		// and serving a landing page instead.
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})
}

// PerformAction is called whenever an action is triggered via the connctd platform.
// It will validate the action request and delegate valid requests to the service.
// If the action is pending, the service should respond with an ActionResponse.
// It is then the responsibility of the service to update the action request state as soon as the request is completed.
// If the action is successfully completed, the service should return nil.
// In case of an error, the service should respond with an appropriate error from errors.go.
func PerformAction(service ConnectorService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ActionRequest
		if err := decodeJSONBody(w, r, &req); err != nil {
			writeError(w, err)
			return
		}

		response, err := service.PerformAction(r.Context(), req)
		if err != nil {
			writeStatus(w, err)
			if response != nil {
				b, err := json.Marshal(response)
				if err != nil {
					writeError(w, err)
					return
				}
				// Do not set a status code, since we want to keep the status set by the error.
				w.Write(b)
			}
			return
		}

		if response != nil {
			if response.ID == "" {
				response.ID = req.ID
			}
			b, err := json.Marshal(response)
			if err != nil {
				writeError(w, err)
				return
			}
			w.Header().Add("Content-Type", "application/json")
			// TODO: should this be http.StatusAccepted?
			w.WriteHeader(http.StatusOK)
			w.Write(b)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

// helps to decode the request body
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dest interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return ErrorBadContentType
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ErrorBadRequestBody
	}

	if err = json.Unmarshal(body, dest); err != nil {
		return ErrorInvalidJsonBody
	}

	return nil
}

// helps to encode an error
func writeError(w http.ResponseWriter, err error) {
	var e api.Error
	if errors.As(err, e) {
		e.Write(w)
	} else {
		api.NewError(
			"INTERNAL_SERVER_ERROR",
			err.Error(),
			http.StatusInternalServerError,
		).Write(w)
	}
}

// helps to set the status according to an error
func writeStatus(w http.ResponseWriter, err error) {
	var e api.Error
	if errors.As(err, e) {
		w.WriteHeader(e.Status)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
