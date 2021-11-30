package giphy

type Provider interface {
	UpdateChannel() <-chan GiphyUpdate
	RegisterInstances(instances ...*Instance) error
	RegisterInstallations(installations ...*Installation) error
}

type GiphyUpdate struct {
	InstanceId string
	Value      string
}
