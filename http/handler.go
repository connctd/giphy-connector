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
		if err = json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		response, err := service.AddInstallation(r.Context(), req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Installation was successfull so far but we require further steps
		if response != nil {
			// We add the further steps and details to the response
			// We set the status code to accepted
			// This signals the connctd platform, that the installation is ongoing
			w.WriteHeader(http.StatusAccepted)
			// As a good citizen we also set an appropriate content type
			w.Header().Add("Content-Type", "application/json")

			b, err := json.Marshal(response)
			if err != nil {
				// Abort if we cannot marshal the response
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("{\"err\":\"Failed to marshal error\"}"))
			} else {
				w.Write(b)
			}
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
		if err = json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := service.AddInstance(r.Context(), req); err != nil {
			// Something went wrong on our side
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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
		if err = json.Unmarshal(body, &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		response, err := service.HandleAction(r.Context(), req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// If the connector service returns with an action response, the action is still pending or it failed.
		// In that case the protocol expects us to send the action response.
		if response != nil {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "application/json")

			b, err := json.Marshal(response)
			if err != nil {
				// Abort if we cannot marshal the response
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("{\"err\":\"Failed to marshal error\"}"))
			} else {
				w.Write(b)
			}
			return
		}

		// Action is complete, we send the appropriate status code according to the connector protocol
		w.WriteHeader(http.StatusNoContent)
	})
}
