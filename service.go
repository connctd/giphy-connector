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
	RemoveInstallation(ctx context.Context, installationId string) error

	AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error
	RemoveInstance(ctx context.Context, installationId string) error

	HandleAction(ctx context.Context, actionRequest connector.ActionRequest) (*connector.ActionResponse, error)
}

type Connector interface {
	CreateThing(ctx context.Context, instanceId string) error
	UpdateProperty(ctx context.Context, thingId string, componentId string, propertyId string, value string) error
}

type Database interface {
	AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) error
	AddInstallationConfiguration(ctx context.Context, installationId string, config []connector.Configuration) error
	GetInstallations(ctx context.Context) ([]*Installation, error)
	RemoveInstallation(ctx context.Context, installationId string) error

	AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error
	GetInstance(ctx context.Context, instanceId string) (*Instance, error)
	GetInstances(ctx context.Context) ([]*Instance, error)
	GetInstanceByThingId(ctx context.Context, thingId string) (*Instance, error)
	RemoveInstance(ctx context.Context, instanceId string) error

	AddThingID(ctx context.Context, instanceID string, thingID string) error
}
