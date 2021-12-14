package connector

import (
	"encoding/json"
	"time"

	"github.com/connctd/restapi-go"
)

// Possible step types:
const (
	StepText     StepType = 1
	StepMarkdown StepType = 2
	StepRedirect StepType = 3
)

// InstallationState reflects the current state of an installation.
type InstallationState int

// Valid installations states:
const (
	InstallationStateInitialized InstallationState = 1
	InstallationStateComplete    InstallationState = 2
	InstallationStateOngoing     InstallationState = 3
	InstallationStateFailed      InstallationState = 4
)

// InstallationRequest sent by connctd in order to signalise a new installation.
type InstallationRequest struct {
	ID            string            `json:"id"`
	Token         InstallationToken `json:"token"`
	State         InstallationState `json:"state"`
	Configuration []Configuration   `json:"configuration"`
}

// InstallationStateUpdateRequest can be sent by a connector to indicate new state.
type InstallationStateUpdateRequest struct {
	State   InstallationState `json:"state"`
	Details json.RawMessage   `json:"details,omitempty"`
}

// InstallationResponse defines the optional response to an installation request.
type InstallationResponse struct {
	Details     json.RawMessage `json:"details,omitempty"`
	FurtherStep Step            `json:"furtherStep,omitempty"`
}

// InstantiationState reflects the current state of an instantiation.
type InstantiationState int

// Valid instantiations states:
const (
	InstantiationStateInitialized InstantiationState = 1
	InstantiationStateComplete    InstantiationState = 2
	InstantiationStateOngoing     InstantiationState = 3
	InstantiationStateFailed      InstantiationState = 4
)

// InstantiationRequest sent by connctd in order to signalise a new instantiation.
type InstantiationRequest struct {
	ID             string             `json:"id"`
	InstallationID string             `json:"installation_id"`
	Token          InstantiationToken `json:"token"`
	State          InstantiationState `json:"state"`
	Configuration  []Configuration    `json:"configuration"`
}

// InstantiationResponse defines the optional response to an instantiation request.
type InstantiationResponse struct {
	Details     json.RawMessage `json:"details,omitempty"`
	FurtherStep Step            `json:"furtherStep,omitempty"`
}

// InstanceStateUpdateRequest can be sent by a connector to indicate a new state.
type InstanceStateUpdateRequest struct {
	State   InstantiationState `json:"state"`
	Details json.RawMessage    `json:"details,omitempty"`
}

// StepType defines the type of a further installation or instantiation step.
type StepType int

// Step defines a further installation or instantiation step
type Step struct {
	Type    StepType `json:"type"`
	Content string   `json:"content"`
}

// AddThingRequest is used to create a new thing on the connctd platform.
type AddThingRequest struct {
	Thing restapi.Thing `json:"thing"`
}

// AddThingResponse describes the response sent by connctd when thing creation was successful.
type AddThingResponse struct {
	ID string `json:"id"`
}

// UpdateThingPropertyValueRequest can be used to propagate a new property value.
type UpdateThingPropertyValueRequest struct {
	Value      string    `json:"value"`
	LastUpdate time.Time `json:"lastUpdate"`
}

// UpdateThingStatusRequest allows updating the status of a thing.
type UpdateThingStatusRequest struct {
	Status restapi.StatusType `json:"status"`
}

// ActionRequest is sent by connctd platform in order to trigger an action.
type ActionRequest struct {
	ID          string                      `json:"id"`
	ThingID     string                      `json:"thingId"`
	ComponentID string                      `json:"componentId"`
	ActionID    string                      `json:"actionId"`
	Status      restapi.ActionRequestStatus `json:"status"`
	Parameters  map[string]string           `json:"parameters"`
}

// ActionResponse can be sent in order to inform about the state of an action.
type ActionResponse struct {
	ID     string                      `json:"id"`
	Status restapi.ActionRequestStatus `json:"status"`
	Error  string                      `json:"error"`
}

// ActionRequestStatusUpdate allows a connector to update the status of an action.
type ActionRequestStatusUpdate struct {
	Status restapi.ActionRequestStatus `json:"status"`
	Error  string                      `json:"error"`
}
