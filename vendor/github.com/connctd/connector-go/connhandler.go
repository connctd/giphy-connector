package connector

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

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

	c.router.Path("/instantiations").Methods(http.MethodPost).Handler(NewSignatureValidationHandler(
		AutoProxyRequestValidationPreProcessor(), publicKey, AddInstance(c.service)))
	c.router.Path("/instantiations/{id}").Methods(http.MethodDelete).Handler(NewSignatureValidationHandler(
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
// I will validate the request and delegate valid requests to the service.
// If the installation needs further steps, the service should respond with an InstallationResponse.
// It is then the responsibility of the service to update the installation state as soon as the installation is completed.
// If the installation is successfully completed, the service should return nil.
// In case of an error, the service should respond with an appropriate error from errors.go.
func AddInstallation(service ConnectorService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var req InstallationRequest
		if err := decodeJSONBody(w, r, &req); err != nil {
			writeError(w, err)
			return
		}

		response, err := service.AddInstallation(r.Context(), req)
		if err != nil {
			writeError(w, err)
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

		w.WriteHeader(http.StatusCreated)
	})
}

//RemoveInstallation is called whenever an installation is removed by the the connctd platform.
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

		w.WriteHeader(http.StatusNoContent)
	})
}

// AddInstance is called whenever a connector is instantiated via the connctd platform.
// I will validate the request and delegate valid requests to the service.
// If the instantiation needs further steps, the service should respond with an InstantiationResponse.
// It is then the responsibility of the service to update the instantiation state as soon as the instantiation is completed.
// If the instantiation is successfully completed, the service should return nil.
// In case of an error, the service should respond with an appropriate error from errors.go.
func AddInstance(service ConnectorService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req InstantiationRequest
		if err := decodeJSONBody(w, r, &req); err != nil {
			writeError(w, err)
			return
		}

		response, err := service.AddInstance(r.Context(), req)
		if err != nil {
			writeError(w, err)
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

		w.WriteHeader(http.StatusNoContent)
	})
}

// PerformAction is called whenever an action is triggered via the connctd platform.
// I will validate the action request and delegate valid requests to the service.
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
			writeError(w, err)
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
	var e Error
	if errors.As(err, &e) {
		e.WriteBody(w)
	} else {
		e = Error{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}

		e.WriteBody(w)
	}
}
