package connector

import (
	"context"
)

// The Provider interface is used in the default service to implement all technology specific details.
// It is the only interface most connectors should need to implement.
// A default implementation meant to be embedded in new connectors is provided by the SDK.
// It implements registration and removal of installations and instances, as well as an update channel and asynchronous action request.
// Connectors can use the update channel to push thing and action request updates to the connctd platform and should implemnet an action handler if they implement actions.
type Provider interface {
	// UpdateChannel is used by the connector service to receive update event.
	// The provider can push updates to the underlying channel or use directly use the connctd API client.
	UpdateChannel() <-chan UpdateEvent

	// RequestAction is called by the connector service when it received an action request.
	// The provider can execute the action synchronously and return an ActionRequestStatusCompleted or ActionRequestStatusFailed.
	// If it returns ActionRequestStatusFailed it is expected to also return an error with details on the failing condition.
	// In both cases an appropriate connector.ActionResponse is returned to the platform.
	// The provider can also decide to execute the action request asynchronously and return an ActionRequestStatusPending.
	// It is then the responsibility of the provider to update the action request as soon as it is finished.
	RequestAction(ctx context.Context, instance *Instance, actionRequest ActionRequest) (ActionRequestStatus, error)

	// RegisterInstallations is called by the connector service to register new installations
	// Installations are registered whenever the service received an successful installation request or when the connector is started.
	RegisterInstallations(installations ...*Installation) error

	// RemoveInstance is called by the service if it received an installation removal request.
	RemoveInstallation(installationId string) error

	// RegisterInstances is called by the connector service to register new instances.
	// Instances are registered whenever the service received an successful instantiation request or when the connector is started.
	RegisterInstances(instances ...*Instance) error

	// RemoveInstance is called by the service if it received an instance removal request.
	RemoveInstance(instanceId string) error
}

// UpdateEvents are pushed to the UpdateChannel.
// The default service will listen to the channel.
// If it receives an UpdateEvent with only a PropertyEventUpdate it will update the specified property with the new value.
// If it receives an ActionEvent it will update the the state of the specified action request to the state in the ActionResponse.
// If the same UpdateEvent contains a PropertyUpateEvent it will first update the property and then the action request.
// If the property update fails it will set the action request state to failed.
type UpdateEvent struct {
	ActionEvent         *ActionEvent
	PropertyUpdateEvent *PropertyUpdateEvent
}

// ActionEvent is used to propagate action request results to the service.
// See UpdateEvent for details.
type ActionEvent struct {
	InstanceId string
	RequestId  string
	Response   *ActionResponse
}

// ActionEvent is used to propagate property updates to the service.
// See UpdateEvent for details.
type PropertyUpdateEvent struct {
	ThingId     string
	InstanceId  string
	ComponentId string
	PropertyId  string
	Value       string
}
