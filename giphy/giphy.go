package giphyprovider

import (
	"errors"
	"time"

	connector "giphy-connector"

	giphyClient "github.com/peterhellberg/giphy"
	"github.com/sirupsen/logrus"
)

var (
	updateChannelBufferSize = 5
)

type Handler struct {
	updateChannel chan connector.GiphyUpdate
	instances     []*connector.Instance
	newInstances  []*connector.Instance
	installations map[string]*connector.Installation
	giphyClient   *giphyClient.Client
}

func New() *Handler {
	client := giphyClient.DefaultClient

	return &Handler{
		instances:     []*connector.Instance{},
		newInstances:  []*connector.Instance{},
		installations: make(map[string]*connector.Installation),
		updateChannel: make(chan connector.GiphyUpdate, updateChannelBufferSize),
		giphyClient:   client,
	}
}

func (h *Handler) UpdateChannel() <-chan connector.GiphyUpdate {
	return h.updateChannel
}

func (h *Handler) RegisterInstances(instances ...*connector.Instance) error {
	h.newInstances = append(h.newInstances, instances...)
	return nil
}

func (h *Handler) RegisterInstallations(installations ...*connector.Installation) error {
	for _, installation := range installations {
		h.installations[installation.ID] = installation
	}
	return nil
}

func (h *Handler) Run() {
	for {
		h.addNewInstances()

		for _, instance := range h.instances {
			if err := h.setApiKey(instance.InstallationID); err != nil {
				logrus.WithError(err).Errorln("failed to set API key for " + instance.InstallationID)
				continue
			}

			randomGif, err := h.getRandomGif()
			if err != nil {
				continue
			}

			h.updateChannel <- connector.GiphyUpdate{
				InstanceId: instance.ID,
				Value:      randomGif,
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func (h *Handler) addNewInstances() {
	h.instances = append(h.instances, h.newInstances...)
	h.newInstances = nil
}

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

func (h *Handler) getRandomGif() (string, error) {
	random, err := h.giphyClient.Random([]string{})
	logrus.WithField("random", random.Data).WithField("url", random.Data.URL).Info("Retrieved Giphy update")
	if err != nil {
		logrus.WithError(err).Errorln("Failed to resolve random gif")
		return "", err
	}
	return random.Data.URL, nil
}
