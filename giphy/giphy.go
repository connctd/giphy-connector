package giphyprovider

import (
	"context"
	"errors"
	"sync"
	"time"

	connector "giphy-connector"

	sdk "github.com/connctd/connector-go"

	"github.com/connctd/restapi-go"
	giphyClient "github.com/peterhellberg/giphy"
	"github.com/sirupsen/logrus"
)

var (
	updateChannelBufferSize = 5
	actionChannelBufferSize = 5
)

type Handler struct {
	updateChannel chan connector.GiphyUpdate
	actionChannel chan pendingAction
	instances     []*connector.Instance
	newInstances  []*connector.Instance
	installations map[string]*connector.Installation
	giphyClient   *giphyClient.Client
	clientLock    sync.Mutex
}

type pendingAction struct {
	sdk.ActionRequest
	instance *connector.Instance
}

// New return a new Giphy provider.
func New() *Handler {
	client := giphyClient.DefaultClient

	return &Handler{
		instances:     []*connector.Instance{},
		newInstances:  []*connector.Instance{},
		installations: make(map[string]*connector.Installation),
		updateChannel: make(chan connector.GiphyUpdate, updateChannelBufferSize),
		actionChannel: make(chan pendingAction, actionChannelBufferSize),
		giphyClient:   client,
	}
}

// UpdateChannel returns the update channel and allows the connector to listen for update events.
func (h *Handler) UpdateChannel() <-chan connector.GiphyUpdate {
	return h.updateChannel
}

// RegisterInstances allows the connector to register instances with the provider.
// Each instance will be periodically updated its random component.
func (h *Handler) RegisterInstances(instances ...*connector.Instance) error {
	h.newInstances = append(h.newInstances, instances...)
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

func (h *Handler) RequestAction(ctx context.Context, instance *connector.Instance, actionRequest sdk.ActionRequest) (restapi.ActionRequestStatus, error) {
	switch actionRequest.ActionID {
	case "search":
		h.actionChannel <- pendingAction{actionRequest, instance}
		return restapi.ActionRequestStatusPending, nil
	default:
		return restapi.ActionRequestStatusFailed, errors.New("action not supported")
	}
}

// Run starts an endless loop which will periodically update the random component of each instance
func (h *Handler) Run() {
	go h.periodicUpdate()
	go h.actionHandler()
}

func (h *Handler) periodicUpdate() {
	for {
		h.addNewInstances()

		for _, instance := range h.instances {
			randomGif, err := h.getRandomGif(instance)
			if err != nil {
				continue
			}

			h.updateChannel <- connector.GiphyUpdate{
				InstanceId:  instance.ID,
				ComponentId: connector.RandomComponentId,
				PropertyId:  connector.RandomPropertyId,
				Value:       randomGif,
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func (h *Handler) actionHandler() {
	for pendingAction := range h.actionChannel {
		switch pendingAction.ActionID {
		case "search":
			update := connector.GiphyUpdate{
				ActionResponse: &sdk.ActionResponse{
					ID: pendingAction.ID,
				},
				InstanceId:  pendingAction.instance.ID,
				ComponentId: connector.SearchComponentId,
				PropertyId:  connector.SearchPropertyId,
				Value:       "",
			}
			keyword := pendingAction.Parameters["keyword"]
			result, err := h.getSearchResult(pendingAction.instance, keyword)
			if err != nil {
				update.ActionResponse.Status = restapi.ActionRequestStatusFailed
				update.ActionResponse.Error = err.Error()
				h.updateChannel <- update
			} else {
				update.ActionResponse.Status = restapi.ActionRequestStatusCompleted
				update.Value = result
				h.updateChannel <- update
			}
		}
	}
}

func (h *Handler) addNewInstances() {
	h.instances = append(h.instances, h.newInstances...)
	h.newInstances = nil
}

// setApiKey will set the Giphy API key to the one configured for installation with the given ID.
// It returns an error if either the installation is not registered or has no API key configuration parameter.
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
