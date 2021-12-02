package main

import (
	"context"
	"giphy-connector"

	"github.com/connctd/connector-go"
	"github.com/connctd/restapi-go"
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

// init is called once during startup of the connector.
// It will register existing installations and instances with the Giphy provider
// and start an event handler to handle events from the provider.
func (g *GiphyConnector) init() {
	installations, err := g.db.GetInstallations(context.Background())
	if err != nil {
		g.logger.WithError(err).Error("Failed to retrieve instances from db")
		return
	}
	g.giphyProvider.RegisterInstallations(installations...)

	instances, err := g.db.GetInstances(context.Background())
	if err != nil {
		g.logger.WithError(err).Error("Failed to retrieve instances from db")
		return
	}

	g.giphyProvider.RegisterInstances(instances...)

	go g.giphyEventHandler(context.Background())
}

// AddInstallation is called by the HTTP handler when it retrieved an installation request
// It will persist the new installation and its configuration and register the new installation with the Giphy provider.
func (g *GiphyConnector) AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) (*connector.InstallationResponse, error) {
	g.logger.WithField("installationRequest", installationRequest).Infoln("Received an installation request")

	if err := g.db.AddInstallation(ctx, installationRequest); err != nil {
		g.logger.WithError(err).Errorln("Failed to add installation")
		return nil, err
	}

	if err := g.db.AddInstallationConfiguration(ctx, installationRequest.ID, installationRequest.Configuration); err != nil {
		g.logger.WithError(err).WithField("config", installationRequest.Configuration).Errorln("Failed to add installation configuration")
		return nil, err
	}

	g.giphyProvider.RegisterInstallations(&giphy.Installation{
		ID:            installationRequest.ID,
		Token:         installationRequest.Token,
		Configuration: installationRequest.Configuration,
	})

	return nil, nil
}

// RemoveInstallation is called by the HTTP handler when it retrieved an installation removal request.
// It will remove the installation from the database (including the installation token) and from the running Giphy provider.
// Note that we will not be able to communicate with the connctd platform about the removed installation after this, since the token is deleted.
func (g *GiphyConnector) RemoveInstallation(ctx context.Context, installationId string) error {
	g.logger.WithField("installationId", installationId).Infoln("Received an installation removal request")

	if err := g.giphyProvider.RemoveInstallation(installationId); err != nil {
		return err
	}

	if err := g.db.RemoveInstallation(ctx, installationId); err != nil {
		return err
	}
	return nil
}

// AddInstantiation is called by the HTTP handler when it retrieved an instantiation request
// It will persist the new instance, create a new Thing for the instance
// and register the new instance with the Giphy provider.
func (g *GiphyConnector) AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error {
	g.logger.WithField("instantiationRequest", instantiationRequest).Infoln("Received an instantiation request")

	if err := g.db.AddInstance(ctx, instantiationRequest); err != nil {
		g.logger.WithError(err).Errorln("Failed to add instance")
		return err
	}

	if err := g.CreateThing(ctx, instantiationRequest.ID); err != nil {
		g.logger.WithError(err).Errorln("Failed to create new thing")
		return err
	}

	g.giphyProvider.RegisterInstances(&giphy.Instance{
		ID:             instantiationRequest.ID,
		InstallationID: instantiationRequest.InstallationID,
		Token:          instantiationRequest.Token,
	})

	return nil
}

// RemoveInstance is called by the HTTP handler when it retrieved an instance removal request.
// It will remove the instance from the database (including the instance token) and from the running Giphy provider.
// Note that we will not be able to communicate with the connctd platform about the removed instance after this, since the token is deleted.
func (g *GiphyConnector) RemoveInstance(ctx context.Context, installationId string) error {
	g.logger.WithField("installationId", installationId).Infoln("Received an installation removal request")

	if err := g.giphyProvider.RemoveInstance(installationId); err != nil {
		return err
	}

	if err := g.db.RemoveInstance(ctx, installationId); err != nil {
		return err
	}
	return nil
}

// HandleAction is called by the HTTP handler when it retrieved an action request
func (g *GiphyConnector) HandleAction(ctx context.Context, actionRequest connector.ActionRequest) (*connector.ActionResponse, error) {
	g.logger.WithField("actionRequest", actionRequest).Infoln("Received an action request")

	instance, err := g.db.GetInstanceByThingId(ctx, actionRequest.ThingID)
	if err != nil {
		g.logger.WithField("actionRequest", actionRequest).WithError(err).Error("Could not retrieve the instance for thing ID")
		return &connector.ActionResponse{Status: restapi.ActionRequestStatusFailed, Error: "thing ID was not found at connector"}, nil
	}

	status, err := g.giphyProvider.RequestAction(ctx, instance, actionRequest)
	if err != nil {
		return &connector.ActionResponse{Status: status, Error: err.Error()}, err
	}

	switch status {
	case restapi.ActionRequestStatusCompleted:
		// The action is completed.
		// We send no error and no response body and the handler will return status code 204.
		return nil, nil
	case restapi.ActionRequestStatusPending:
		// The action is not completed yet.
		// We send no error but an ActionResponse and the handler will return status code 200.
		// We have to send an status update when the action is completed.
		return &connector.ActionResponse{Status: status}, nil
	case restapi.ActionRequestStatusFailed:
		// This should not happen.
		// The provider is expected to return an error if the action failed, which we catch above.
		g.logger.WithField("actionRequest", actionRequest).Debug("connector did not send an error but set action state to FAILED")
	}

	return nil, nil
}

// giphyEventHandler handles events coming from the giphy provider
func (g *GiphyConnector) giphyEventHandler(ctx context.Context) {
	// wait for Giphy events
	// TODO: Remove goroutine. the event handler is already run as a goroutine
	go func() {
		for update := range g.giphyProvider.UpdateChannel() {
			g.logger.WithField("update", update).WithField("value", update.Value).Infoln("Received update from Giphy provider")
			g.UpdateProperty(ctx, update.InstanceId, update.ComponentId, update.PropertyId, update.Value)
			if update.ActionResponse != nil {
				g.logger.WithField("action", update.ActionResponse).Info("update action status")
				if update.ActionRequestId == "" {
					g.logger.WithField("update", update).Errorln("no action request id set")
					continue
				}
				err := g.UpdateActionStatus(ctx, update.InstanceId, update.ActionRequestId, update.ActionResponse)
				if err != nil {
					g.logger.WithError(err).Error("Failed to update action status")
				}
			}
		}
	}()
}
