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
	logger        logrus.FieldLogger
	db            giphy.Database
	connctdClient connector.Client
	giphyProvider giphy.Provider
}

// NewService returns a new instance of the Giphy connector
func NewService(dbClient giphy.Database, connctdClient connector.Client, giphyProvider giphy.Provider, logger logrus.FieldLogger) giphy.Service {
	connector := &GiphyConnector{
		logger,
		dbClient,
		connctdClient,
		giphyProvider,
	}

	connector.init()

	return connector
}

func (g *GiphyConnector) init() {
	instances, err := g.db.GetInstances(context.Background())
	if err != nil {
		g.logger.WithError(err).Error("Failed to retrieve instances from db")
		return
	}

	instanceIds := make([]string, len(instances))
	for i := range instances {
		instanceIds[i] = instances[i].ID
	}
	g.giphyProvider.RegisterInstances(instanceIds...)

	go g.giphyEventHandler(context.Background())
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

	if err := g.CreateThing(ctx, instantiationRequest.ID); err != nil {
		g.logger.WithError(err).Errorln("Failed to create new thing")
		return err
	}

	g.giphyProvider.RegisterInstances(instantiationRequest.ID)

	return nil
}

// HandleAction is called by the HTTP handler when it retrieved an action request
func (g *GiphyConnector) HandleAction(ctx context.Context, actionRequest connector.ActionRequest) error {
	logrus.WithField("actionRequest", actionRequest).Infoln("Received an action request")
	return nil
}

// giphyEventHandler handles events coming from the giphy provider
func (g *GiphyConnector) giphyEventHandler(ctx context.Context) {
	// wait for Giphy events
	go func() {
		for update := range g.giphyProvider.UpdateChannel() {
			g.logger.WithField("value", update.Value).Infoln("Received update from Giphy provider")
			g.UpdateProperty(ctx, update.InstanceId, update.Value)
		}
	}()
}
