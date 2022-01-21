// Package connector implements a SDK for connector development.
// In most cases it should be sufficient to use the default provider to implement the connector.Provider interface and use it with the default service and the connector handler.
// For an example connector using this SDK to integrate an external API, see https://github.com/connctd/giphy-connector/.
package connector

import (
	"context"

	"github.com/connctd/connector-go/connctd"
)

// ConnectorService interface is used by the ConnectorHandler and will be called to process the validated requests used in the connector protocol.
// The SDK provides a default implementation for the ConnectorService interface that should be sufficient for most connector developments.
type ConnectorService interface {
	// AddInstallation is called by the ConnectorHandler when it received an installation request.
	// The request is validated before calling AddInstallation but the connector can implemnet additional validation.
	// If the installation is completed successfully, the service should return nil and no error.
	// If the installation needs further steps, the service should respond with an InstallationResponse and no error.
	// It is then the responsibility of the service to update the installation state as soon as the installation is completed.
	// In case of an error, the service should respond with an appropriate error from errors.go and can also return an InstallationResponse.
	// The status code will be set to one defined in the error and the InstallationResponse will be returned to the connctd platform.
	AddInstallation(ctx context.Context, request InstallationRequest) (*InstallationResponse, error)

	// RemoveInstallation is called whenever an installation is removed by the the connctd platform.
	// The connector should remove the installation and can return an error if needed.
	// Regardless of the return value, the installation is removed from the connctd platform.
	RemoveInstallation(ctx context.Context, installationId string) error

	// AddInstance is called by the ConnectorHandler whenever a connector is instantiated via the connctd platform.
	// The request is validated before calling AddInstance but the connector can implemnet additional validation.
	// If the instantiation is completed successfully, the service should return nil and no error.
	// If the instantiation needs further steps, the service should respond with an InstantiationResponse and no error.
	// It is then the responsibility of the service to update the instantiation state as soon as the instantiation is completed.
	// In case of an error, the service should respond with an appropriate error from errors.go and can also return an InstantiationResponse.
	// The status code will be set to one defined in the error and the InstantiationResponse will be returned to the connctd platform.
	AddInstance(ctx context.Context, request InstantiationRequest) (*InstantiationResponse, error)

	//RemoveInstance is called whenever an instance is removed by the the connctd platform.
	RemoveInstance(ctx context.Context, instanceId string) error

	// PerformAction is called by the ConnectorHandler whenever an action is triggered via the connctd platform.
	// The request is validated before calling PerformAction but the connector can implement additional validation.
	// If the action is pending, the service should respond with an ActionResponse.
	// It is then the responsibility of the service to update the action request state as soon as the request is completed.
	// If the action is successfully completed, the service should return nil.
	// In case of an error, the service should respond with an appropriate error from errors.go.
	PerformAction(ctx context.Context, request ActionRequest) (*ActionResponse, error)
}

// ThingTemplate describes the thing together with an external ID that is created for each new instance.
// If the connector doesn't need an external ID it can be left blank.
type ThingTemplate struct {
	Thing      connctd.Thing
	ExternalID string
}

// ThingTemplates is used by the default connector service to create a set of connctd.Thing for each new instantiation request.
// For each template in the returned slice, the default service will:
// - create the connctd.Thing with the connctd platform
// - store a connector.ThingMapping{} with the instantiation request ID, the ID of the created thing and the external ID
// - register the new thing with the provider
type ThingTemplates func(request InstantiationRequest) []ThingTemplate

// Database interface is used in the default service to persist new installations, instances, configurations and external device mappings.
// The SDK provides a default implementation supporting Postgresql, Mysql and Sqlite3.
type Database interface {
	AddInstallation(ctx context.Context, installationRequest InstallationRequest) error
	AddInstallationConfiguration(ctx context.Context, installationId string, config []Configuration) error
	GetInstallations(ctx context.Context) ([]*Installation, error)
	RemoveInstallation(ctx context.Context, installationId string) error

	AddInstance(ctx context.Context, instantiationRequest InstantiationRequest) error
	AddInstanceConfiguration(ctx context.Context, instanceId string, config []Configuration) error
	GetInstance(ctx context.Context, instanceId string) (*Instance, error)
	GetInstances(ctx context.Context) ([]*Instance, error)
	GetInstanceByThingId(ctx context.Context, thingId string) (*Instance, error)
	GetInstanceConfiguration(ctx context.Context, instanceId string) ([]Configuration, error)
	GetMappingByInstanceId(ctx context.Context, instanceId string) ([]ThingMapping, error)
	RemoveInstance(ctx context.Context, instanceId string) error

	AddThingMapping(ctx context.Context, instanceID string, thingID string, externalId string) error
}
