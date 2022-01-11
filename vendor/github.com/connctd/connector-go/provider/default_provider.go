// Package provider implements the basic bookeeping needed by most providers.
// It is meant to be embedded by provider implementations.
package provider

import (
	"context"
	"errors"

	"github.com/connctd/connector-go"
)

var (
	updateChannelBufferSize = 5
	actionChannelBufferSize = 5
)

type DefaultProvider struct {
	Installations         map[string]*connector.Installation
	Instances             []*connector.Instance
	actionChannel         chan PendingAction
	updateChannel         chan connector.UpdateEvent
	newInstances          []*connector.Instance
	instancesToRemove     []string
	newInstallations      []*connector.Installation
	installationsToRemove []string
}

func New() DefaultProvider {
	return DefaultProvider{
		Installations:    make(map[string]*connector.Installation),
		Instances:        []*connector.Instance{},
		newInstances:     []*connector.Instance{},
		newInstallations: []*connector.Installation{},
		updateChannel:    make(chan connector.UpdateEvent, updateChannelBufferSize),
		actionChannel:    make(chan PendingAction, actionChannelBufferSize),
	}
}

type PendingAction struct {
	connector.ActionRequest
	Instance *connector.Instance
}

// UpdateChannel returns the update channel and allows the connector to listen for update events.
func (p *DefaultProvider) UpdateChannel() <-chan connector.UpdateEvent {
	return p.updateChannel
}

// UpdateEvent publishes the update event on the update event channel.
func (p *DefaultProvider) UpdateEvent(update connector.UpdateEvent) {
	p.updateChannel <- update
}

// UpdateChannel returns the action channel and allows the provider to listen for action events.
func (p *DefaultProvider) ActionChannel() <-chan PendingAction {
	return p.actionChannel
}

// ActionEvent publishes the action event on the action event channel.
// This can be used to implement asynchronous RequestAction() method in the provider.
func (p *DefaultProvider) ActionEvent(action PendingAction) {
	p.actionChannel <- action
}

// RegisterInstances allows the connector to register instances with the provider.
// Each instance will be periodically updated its random component.
func (p *DefaultProvider) RegisterInstances(instances ...*connector.Instance) error {
	p.newInstances = append(p.newInstances, instances...)
	return nil
}

// RemoveInstance marks the instance with the given id for removal.
// The instance will be removed before the next run of the periodic update.
// If it is not registered, RemoveInstance will return an error.
func (p *DefaultProvider) RemoveInstance(instanceId string) error {
	index := findIndex(p.Instances, instanceId)

	if index > -1 {
		p.instancesToRemove = append(p.instancesToRemove, instanceId)
	} else {
		return errors.New("instance not found")
	}

	return nil
}

// RegisterInstallations allows the connector to register new installations with the provider.
// The provider needs access to the installation in order to use installation specific configuration parameters.
func (p *DefaultProvider) RegisterInstallations(installations ...*connector.Installation) error {
	p.newInstallations = append(p.newInstallations, installations...)
	return nil
}

// RemoveInstallation removes the installation with the given id from the provider.
func (p *DefaultProvider) RemoveInstallation(installationId string) error {
	_, ok := p.Installations[installationId]
	if !ok {
		return errors.New("installation not found")
	}

	p.installationsToRemove = append(p.installationsToRemove, installationId)

	return nil
}

// RequestAction allows the connector to send a action request to the Giphy provider.
// All actions are executed asynchronously by sending it to the actions channel.
// The provider implementation is responsible for listening to that channel and sending an appropriate response
// to the connector using the updates channel.
// Synchronous actions can be implemented by overriding this method and returning a non pending action request status.
func (p *DefaultProvider) RequestAction(ctx context.Context, instance *connector.Instance, actionRequest connector.ActionRequest) (connector.ActionRequestStatus, error) {
	p.ActionEvent(PendingAction{actionRequest, instance})
	return connector.ActionRequestStatusPending, nil
}

// AddNewInstallations will add all newly registered installations to p.Instances.
// The provider is expected to call this to be able to use newly registered installations.
func (p *DefaultProvider) AddNewInstallations() {
	for _, installation := range p.newInstallations {
		p.Installations[installation.ID] = installation
	}
}

// RemoveInstallations removes all installations that are marked for removal.
// It ignores installations that are not found.
// The provider is expected to call this before using p.Installations.
func (p *DefaultProvider) RemoveInstallations() {
	for _, installationId := range p.installationsToRemove {
		delete(p.Installations, installationId)
	}

	p.installationsToRemove = nil
}

// RemoveInstances removes all instances that are marked for removal.
// It ignores instances that are not found.
// The provider is expected to call this before using p.Instances
func (p *DefaultProvider) RemoveInstances() {
	for i := range p.instancesToRemove {
		index := findIndex(p.Instances, p.instancesToRemove[i])

		if index > -1 {
			p.Instances = remove(p.Instances, index)
		}
	}

	p.instancesToRemove = nil
}

// AddNewInstances will add all newly registered instances.
// The provider is expected to call this to be able to use newly registered instances.
func (p *DefaultProvider) AddNewInstances() {
	p.Instances = append(p.Instances, p.newInstances...)
	p.newInstances = nil
}

// Update will add all newly registered installations and instances and will remove the installations and instances that are marked for removal.
func (p *DefaultProvider) Update() {
	p.RemoveInstances()
	p.AddNewInstances()
	p.RemoveInstallations()
	p.AddNewInstallations()
}

// findIndex return the index of the instance with the given id.
func findIndex(instances []*connector.Instance, instanceId string) int {
	for i := range instances {
		if instances[i].ID == instanceId {
			return i
		}
	}
	return -1
}

// remove will remove the instance at the given index and return the resulting slice
// It assumes the index to be in range.
func remove(instances []*connector.Instance, index int) []*connector.Instance {
	instances[index] = instances[len(instances)-1]
	return instances[:len(instances)-1]
}
