package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/connctd/connector-go"

	"github.com/connctd/restapi-go"
	giphyClient "github.com/peterhellberg/giphy"
	"github.com/sirupsen/logrus"
)

var (
	updateChannelBufferSize = 5
	actionChannelBufferSize = 5
)

type Handler struct {
	updateChannel     chan connector.UpdateEvent
	actionChannel     chan pendingAction
	instances         []*connector.Instance
	newInstances      []*connector.Instance
	instancesToRemove []string
	installations     map[string]*connector.Installation
	giphyClient       *giphyClient.Client
	clientLock        sync.Mutex
}

type pendingAction struct {
	connector.ActionRequest
	instance *connector.Instance
}

// New return a new Giphy provider.
func NewGiphyProvider() *Handler {
	client := giphyClient.DefaultClient

	return &Handler{
		instances:     []*connector.Instance{},
		newInstances:  []*connector.Instance{},
		installations: make(map[string]*connector.Installation),
		updateChannel: make(chan connector.UpdateEvent, updateChannelBufferSize),
		actionChannel: make(chan pendingAction, actionChannelBufferSize),
		giphyClient:   client,
	}
}

// UpdateChannel returns the update channel and allows the connector to listen for update events.
func (h *Handler) UpdateChannel() <-chan connector.UpdateEvent {
	return h.updateChannel
}

// RegisterInstances allows the connector to register instances with the provider.
// Each instance will be periodically updated its random component.
func (h *Handler) RegisterInstances(instances ...*connector.Instance) error {
	h.newInstances = append(h.newInstances, instances...)
	return nil
}

// RemoveInstance marks the instance with the given id for removal.
// The instance will be removed before the next run of the periodic update.
// If it is not registered, RemoveInstance will return an error.
func (h *Handler) RemoveInstance(instanceId string) error {
	index := findIndex(h.instances, instanceId)

	if index > -1 {
		h.instancesToRemove = append(h.instancesToRemove, instanceId)
	} else {
		return errors.New("instance not found")
	}

	return nil
}

// RegisterInstallations allows the connector to register new installations with the provider.
// The provider needs access to the installation in order to use installation specific configuration parameters.
func (h *Handler) RegisterInstallations(installations ...*connector.Installation) error {
	for _, installation := range installations {
		h.installations[installation.ID] = installation
	}
	return nil
}

// RemoveInstallation removes the installation with the given id from the provider.
func (h *Handler) RemoveInstallation(installationId string) error {
	_, ok := h.installations[installationId]
	if !ok {
		return errors.New("installation not found")
	}

	delete(h.installations, installationId)

	return nil
}

// RequestAction allows the connector to send a action request to the Giphy provider.
// We currently only implement the search request which we will execute asynchronously via the action handler below.
// Other actions could be directly executed and return a appropriate status.
func (h *Handler) RequestAction(ctx context.Context, instance *connector.Instance, actionRequest connector.ActionRequest) (restapi.ActionRequestStatus, error) {
	switch actionRequest.ActionID {
	case "search":
		h.actionChannel <- pendingAction{actionRequest, instance}
		return restapi.ActionRequestStatusPending, nil
	default:
		return restapi.ActionRequestStatusFailed, errors.New("action not supported")
	}
}

// Run starts the periodic update and the action handler.
func (h *Handler) Run() {
	go h.periodicUpdate()
	go h.actionHandler()
}

// periodicUpdate starts an endless loop which will periodically update the random component of each instance
func (h *Handler) periodicUpdate() {
	for {
		h.removeInstances()
		h.addNewInstances()

		for _, instance := range h.instances {
			if len(instance.ThingIDs) <= 0 {
				logrus.WithField("instance", instance).Info("missing thing id")
				continue
			}
			randomGif, err := h.getRandomGif(instance)
			if err != nil {
				continue
			}

			h.updateChannel <- connector.UpdateEvent{
				PropertyUpdateEvent: &connector.PropertyUpdateEvent{
					InstanceId:  instance.ID,
					ThingId:     instance.ThingIDs[0],
					ComponentId: RandomComponentId,
					PropertyId:  RandomPropertyId,
					Value:       randomGif,
				},
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

// actionHandler will listen for and execute action requests
func (h *Handler) actionHandler() {
	for pendingAction := range h.actionChannel {
		switch pendingAction.ActionID {
		case "search":
			update := connector.UpdateEvent{
				ActionEvent: &connector.ActionEvent{
					ActionResponse:  &connector.ActionResponse{},
					ActionRequestId: pendingAction.ID,
					InstanceId:      pendingAction.instance.ID,
				},
			}
			keyword := pendingAction.Parameters["keyword"]
			result, err := h.getSearchResult(pendingAction.instance, keyword)
			if err != nil {
				update.ActionEvent.ActionResponse.Status = restapi.ActionRequestStatusFailed
				update.ActionEvent.ActionResponse.Error = err.Error()
				h.updateChannel <- update
			} else {
				update.ActionEvent.ActionResponse.Status = restapi.ActionRequestStatusCompleted
				update.PropertyUpdateEvent = &connector.PropertyUpdateEvent{
					ThingId:     pendingAction.instance.ThingIDs[0],
					InstanceId:  pendingAction.instance.ID,
					ComponentId: SearchComponentId,
					PropertyId:  SearchPropertyId,
					Value:       result,
				}
				h.updateChannel <- update
			}
		}
	}
}

// setApiKey will set the Giphy API key to the one configured for installation with the given ID.
// It returns an error if either the installation is not registered or has no API key configuration parameter.
// We potentially have multiple goroutines access the Giphy API client and calling this method.
// Therefore we protect it with a mutex which should be locked before calling this.
// See getRandomGif or getSearchResult for details.
// Alternatively we could create and use one API client per installation.
func (h *Handler) setApiKey(installationId string) error {
	installation, ok := h.installations[installationId]
	if !ok {
		return errors.New("installation not registered")
	}
	key, ok := installation.GetConfig("giphy_api_key")
	if !ok {
		return errors.New("could not find api key")
	}

	h.giphyClient.APIKey = key.Value
	return nil
}

// getRandomGif uses the Giphy API to return a new random gif.
func (h *Handler) getRandomGif(instance *connector.Instance) (string, error) {
	h.clientLock.Lock()
	defer h.clientLock.Unlock()
	if err := h.setApiKey(instance.InstallationID); err != nil {
		logrus.WithError(err).Errorln("failed to set API key for " + instance.InstallationID)
		return "", err
	}

	random, err := h.giphyClient.Random([]string{})
	if err != nil {
		logrus.WithError(err).Errorln("Failed to resolve random gif")
		return "", err
	}
	return random.Data.URL, nil
}

// getSearchResult uses the Giphy API to search for the given keyword.
func (h *Handler) getSearchResult(instance *connector.Instance, keyword string) (string, error) {
	h.clientLock.Lock()
	defer h.clientLock.Unlock()
	if err := h.setApiKey(instance.InstallationID); err != nil {
		logrus.WithError(err).Errorln("failed to set API key for " + instance.InstallationID)
		return "", err
	}

	h.giphyClient.Limit = 1
	result, err := h.giphyClient.Search([]string{keyword})
	if err != nil {
		return "", err
	}
	if len(result.Data) <= 0 {
		return "", errors.New("no search result found")
	}

	logrus.WithField("keyword", keyword).WithField("searchResult", result.Data).WithField("url", result.Data[0].URL).Info("Search finished")
	return result.Data[0].URL, nil
}

// removeInstances removes all instances that are marked for removal.
// It ignores instances that are not found.
func (h *Handler) removeInstances() {
	for i := range h.instancesToRemove {
		index := findIndex(h.instances, h.instancesToRemove[i])

		if index > -1 {
			h.instances = remove(h.instances, index)
		}
	}

	h.instancesToRemove = nil
}

// addNewInstances will add all newly registered instances.
func (h *Handler) addNewInstances() {
	h.instances = append(h.instances, h.newInstances...)
	h.newInstances = nil
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
