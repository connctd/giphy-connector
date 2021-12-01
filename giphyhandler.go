package giphy

import (
	"context"

	"github.com/connctd/connector-go"
	"github.com/connctd/restapi-go"
)

type Provider interface {
	UpdateChannel() <-chan GiphyUpdate
	RegisterInstances(instances ...*Instance) error
	RegisterInstallations(installations ...*Installation) error
	RequestAction(ctx context.Context, instance *Instance, actionRequest connector.ActionRequest) (restapi.ActionRequestStatus, error)
}

type GiphyUpdate struct {
	ActionResponse *connector.ActionResponse
	InstanceId     string
	ComponentId    string
	PropertyId     string
	Value          string
}
