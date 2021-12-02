package giphy

import (
	"context"

	"github.com/connctd/connector-go"
	"github.com/connctd/restapi-go"
)

type Provider interface {
	UpdateChannel() <-chan GiphyUpdate
	RegisterInstances(instances ...*Instance) error
	RemoveInstance(instanceId string) error
	RegisterInstallations(installations ...*Installation) error
	RemoveInstallation(installationId string) error
	RequestAction(ctx context.Context, instance *Instance, actionRequest connector.ActionRequest) (restapi.ActionRequestStatus, error)
}

type GiphyUpdate struct {
	ActionResponse *connector.ActionResponse
	InstanceId     string
	ComponentId    string
	PropertyId     string
	Value          string
}
