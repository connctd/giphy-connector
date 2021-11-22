package giphy

import (
	"context"

	"github.com/connctd/connector-go"
)

type Service interface {
	ConnectorProtocol
}

type ConnectorProtocol interface {
	AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) (*connector.InstallationResponse, error)
	AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error
	HandleAction(ctx context.Context, actionRequest connector.ActionRequest) error
}
