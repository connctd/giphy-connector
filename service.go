package giphy

import (
	"context"

	"github.com/connctd/connector-go"
)

type Service interface {
	ConnectorProtocol
	Connector
}

type ConnectorProtocol interface {
	AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) (*connector.InstallationResponse, error)
	AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error
	HandleAction(ctx context.Context, actionRequest connector.ActionRequest) error
}

type Connector interface {
	CreateThing(ctx context.Context, instanceId string) error
	UpdateProperty(ctx context.Context, thingId string, value string) error
}

type Database interface {
	AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) error

	AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error
	GetInstance(ctx context.Context, instanceId string) (*Instance, error)

	AddThingID(ctx context.Context, instanceID string, thingID string) error
}
