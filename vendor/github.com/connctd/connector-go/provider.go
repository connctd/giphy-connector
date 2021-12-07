package connector

import (
	"context"

	"github.com/connctd/restapi-go"
)

type Provider interface {
	UpdateChannel() <-chan UpdateEvent
	RegisterInstances(instances ...*Instance) error
	RemoveInstance(instanceId string) error
	RegisterInstallations(installations ...*Installation) error
	RemoveInstallation(installationId string) error
	RequestAction(ctx context.Context, instance *Instance, actionRequest ActionRequest) (restapi.ActionRequestStatus, error)
}

type UpdateEvent struct {
	ActionEvent         *ActionEvent
	PropertyUpdateEvent *PropertyUpdateEvent
}

type ActionEvent struct {
	InstanceId      string
	ActionResponse  *ActionResponse
	ActionRequestId string
}

type PropertyUpdateEvent struct {
	ThingId     string
	InstanceId  string
	ComponentId string
	PropertyId  string
	Value       string
}
