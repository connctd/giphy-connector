package giphy

type Provider interface {
	UpdateChannel() <-chan GiphyUpdate
	RegisterInstances(instanceId ...string) error
}

type GiphyUpdate struct {
	InstanceId string
	Value      string
}
