package connector

import (
	"context"

	"github.com/connctd/restapi-go"
)

type ConnectorService interface {
	AddInstallation(ctx context.Context, request InstallationRequest) (*InstallationResponse, error)
	RemoveInstallation(ctx context.Context, installationId string) error

	AddInstance(ctx context.Context, request InstantiationRequest) (*InstantiationResponse, error)
	RemoveInstance(ctx context.Context, instanceId string) error

	PerformAction(ctx context.Context, request ActionRequest) (*ActionResponse, error)
}

type ThingService interface {
	CreateThing(ctx context.Context, instanceId string, thing restapi.Thing) (restapi.Thing, error)
	UpdateProperty(ctx context.Context, instanceId, componentId, propertyId, value string) error
	UpdateActionStatus(ctx context.Context, instanceId string, actionRequestId string, actionResponse *ActionResponse) error
}

type ThingTemplates func(request InstantiationRequest) []restapi.Thing
