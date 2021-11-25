package giphyprovider

import (
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
	instances     []string
	newInstances  []string
	giphyClient   *giphyClient.Client
}

func New(apiKey string) *Handler {
	client := giphyClient.DefaultClient
	client.APIKey = apiKey
	return &Handler{
		instances:     []string{},
		newInstances:  []string{},
		updateChannel: make(chan connector.GiphyUpdate, updateChannelBufferSize),
		giphyClient:   client,
	}
}

func (h *Handler) UpdateChannel() <-chan connector.GiphyUpdate {
	return h.updateChannel
}

func (h *Handler) RegisterInstances(instanceIds ...string) error {
	h.newInstances = append(h.newInstances, instanceIds...)
	return nil
}

func (h *Handler) Run() {
	for {
		h.addNewInstances()
		randomGif, err := h.getRandomGif()
		for _, instanceId := range h.instances {
			if err != nil {
				continue
			}
			h.updateChannel <- connector.GiphyUpdate{
				InstanceId: instanceId,
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

func (h *Handler) getRandomGif() (string, error) {
	random, err := h.giphyClient.Random([]string{})
	logrus.WithField("random", random.Data).WithField("url", random.Data.URL).Info("Retrieved Giphy update")
	if err != nil {
		logrus.WithError(err).Errorln("Failed to resolve random gif")
		return "", err
	}
	return random.Data.URL, nil
}
