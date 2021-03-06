package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/connctd/connector-go"
	"github.com/connctd/connector-go/provider"

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
func (h *GiphyProvider) Run(ctx context.Context) {
	go h.periodicUpdate(ctx)
	go h.actionHandler()
}

// periodicUpdate starts an endless loop which will periodically update the random component of each instance
func (h *GiphyProvider) periodicUpdate(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			h.Update()

			for _, instance := range h.Instances {
				if len(instance.ThingMapping) <= 0 {
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
						ThingId:     instance.ThingMapping[0].ThingID,
						ComponentId: RandomComponentId,
						PropertyId:  RandomPropertyId,
						Value:       randomGif,
					},
				}
				h.UpdateEvent(update)
			}
		}
	}
}

// actionHandler will listen for and execute action requests
func (h *GiphyProvider) actionHandler() {
	for pendingAction := range h.ActionChannel() {
		update := connector.UpdateEvent{
			ActionEvent: &connector.ActionEvent{
				InstanceId: pendingAction.Instance.ID,
				RequestId:  pendingAction.ID,
				Response:   &connector.ActionResponse{},
			},
		}

		switch pendingAction.ActionID {
		case "search":
			keyword := pendingAction.Parameters["keyword"]
			result, err := h.getSearchResult(pendingAction.Instance, keyword)

			if err != nil {
				update.ActionEvent.Response = &connector.ActionResponse{
					Status: connector.ActionRequestStatusFailed,
					Error:  err.Error(),
				}
				h.UpdateEvent(update)
				continue
			}

			update.ActionEvent.Response = &connector.ActionResponse{
				Status: connector.ActionRequestStatusCompleted,
			}
			update.PropertyUpdateEvent = &connector.PropertyUpdateEvent{
				ThingId:     pendingAction.Instance.ThingMapping[0].ThingID,
				InstanceId:  pendingAction.Instance.ID,
				ComponentId: SearchComponentId,
				PropertyId:  SearchPropertyId,
				Value:       result,
			}
			h.UpdateEvent(update)

		default:
			update.ActionEvent.Response = &connector.ActionResponse{
				Status: connector.ActionRequestStatusFailed,
				Error:  "Action not supported",
			}
			h.UpdateEvent(update)
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
