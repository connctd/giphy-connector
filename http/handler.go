package http

import (
	"encoding/json"
	"giphy-connector"
	"io/ioutil"
	"net/http"

	"github.com/connctd/connector-go"
)

// handleInstallation is called whenever a connector is installed via the connctd platform
// The request body should contain a valid connector.InstallationRequest
func handleInstallation(service giphy.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req connector.InstallationRequest
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(body, req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Installation was successful and we do not need any further steps
		// No payload is added to the response
		w.WriteHeader(http.StatusCreated)
	})
}

// handleInstantiation is called whenever a connector is instantiated via the connctd platform
// The request body should contain a valid connector.InstantiationRequest
func handleInstantiation(service giphy.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req connector.InstantiationRequest
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(body, req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// TODO: let the connector handle the new instantiation

		w.WriteHeader(http.StatusCreated)
	})
}

// handleAction is called whenever an action is invoked via the connctd platform
// The request body should contain a valid connector.ActionRequest
func handleAction(service giphy.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req connector.ActionRequest
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(body, req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// TODO: let the connector handle action request

		// Action is complete, we send the appropriate status code according to the connector protocol
		w.WriteHeader(http.StatusNoContent)
	})
}
