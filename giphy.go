package main

import (
	"errors"
	"sync"
	"time"

	"github.com/connctd/connector-go"
	"github.com/connctd/connector-go/provider"

	"github.com/connctd/restapi-go"
	giphyClient "github.com/peterhellberg/giphy"
	"github.com/sirupsen/logrus"
)

type GiphyProvider struct {
	provider.DefaultProvider
	giphyClient *giphyClient.Client
	clientLock  sync.Mutex
}

// New return a new Giphy provider.
func NewGiphyProvider() *GiphyProvider {
	client := giphyClient.DefaultClient
	provider := provider.New()

	return &GiphyProvider{
		provider,
		client,
		sync.Mutex{},
	}
}

// Run starts the periodic update and the action handler.
func (h *GiphyProvider) Run() {
	go h.periodicUpdate()
	go h.actionHandler()
}

// periodicUpdate starts an endless loop which will periodically update the random component of each instance
func (h *GiphyProvider) periodicUpdate() {
	for {
		h.RemoveInstances()
		h.AddNewInstances()

		for _, instance := range h.Instances {
			if len(instance.ThingIDs) <= 0 {
				logrus.WithField("instance", instance).Info("missing thing id")
				continue
			}
			randomGif, err := h.getRandomGif(instance)
			if err != nil {
				continue
			}

			update := connector.UpdateEvent{
				PropertyUpdateEvent: &connector.PropertyUpdateEvent{
					InstanceId:  instance.ID,
					ThingId:     instance.ThingIDs[0],
					ComponentId: RandomComponentId,
					PropertyId:  RandomPropertyId,
					Value:       randomGif,
				},
			}
			h.UpdateEvent(update)
		}
		time.Sleep(1 * time.Minute)
	}
}

// actionHandler will listen for and execute action requests
func (h *GiphyProvider) actionHandler() {
	for pendingAction := range h.ActionChannel() {
		switch pendingAction.ActionID {
		case "search":
			update := connector.UpdateEvent{
				ActionEvent: &connector.ActionEvent{
					ActionResponse:  &connector.ActionResponse{},
					ActionRequestId: pendingAction.ID,
					InstanceId:      pendingAction.Instance.ID,
				},
			}
			keyword := pendingAction.Parameters["keyword"]
			result, err := h.getSearchResult(pendingAction.Instance, keyword)
			if err != nil {
				update.ActionEvent.ActionResponse.Status = restapi.ActionRequestStatusFailed
				update.ActionEvent.ActionResponse.Error = err.Error()
				h.UpdateEvent(update)
			} else {
				update.ActionEvent.ActionResponse.Status = restapi.ActionRequestStatusCompleted
				update.PropertyUpdateEvent = &connector.PropertyUpdateEvent{
					ThingId:     pendingAction.Instance.ThingIDs[0],
					InstanceId:  pendingAction.Instance.ID,
					ComponentId: SearchComponentId,
					PropertyId:  SearchPropertyId,
					Value:       result,
				}
				h.UpdateEvent(update)
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
func (h *GiphyProvider) setApiKey(installationId string) error {
	installation, ok := h.Installations[installationId]
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
func (h *GiphyProvider) getRandomGif(instance *connector.Instance) (string, error) {
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
func (h *GiphyProvider) getSearchResult(instance *connector.Instance, keyword string) (string, error) {
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
