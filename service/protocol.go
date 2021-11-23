package main

import (
	"context"
	"giphy-connector"

	"github.com/connctd/connector-go"
	"github.com/sirupsen/logrus"
)

// GiphyConnector provides the callback functions used by the HTTP handler
// Later on it will implement the communication with the Giphy API
type GiphyConnector struct {
	logger logrus.FieldLogger
	db     giphy.Database
}

// NewService returns an new instance of the Giphy connector
func NewService(dbClient giphy.Database, logger logrus.FieldLogger) giphy.Service {
	return &GiphyConnector{
		logger,
		dbClient,
	}
}

// AddInstallation is called by the HTTP handler when it retrieved an installation request
func (g *GiphyConnector) AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) (*connector.InstallationResponse, error) {
	logrus.WithField("installationRequest", installationRequest).Infoln("Received an installation request")

	if err := g.db.AddInstallation(ctx, installationRequest); err != nil {
		g.logger.WithError(err).Errorln("Failed to add installation")
		return nil, err
	}

	return nil, nil
}

// AddInstantiation is called by the HTTP handler when it retrieved an instantiation request
func (g *GiphyConnector) AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error {
	logrus.WithField("instantiationRequest", instantiationRequest).Infoln("Received an instantiation request")

	if err := g.db.AddInstance(ctx, instantiationRequest); err != nil {
		g.logger.WithError(err).Errorln("Failed to add instance")
		return err
	}

	return nil
}

// HandleAction is called by the HTTP handler when it retrieved an action request
func (g *GiphyConnector) HandleAction(ctx context.Context, actionRequest connector.ActionRequest) error {
	logrus.WithField("actionRequest", actionRequest).Infoln("Received an action request")
	return nil
}
