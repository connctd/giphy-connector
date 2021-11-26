package giphy

import "github.com/connctd/connector-go"

type Installation struct {
	ID            string                      `db:"id" json:"id"`
	Token         connector.InstallationToken `db:"token" json:"token"`
	Configuration []connector.Configuration   `json:"configuration"`
}

//GetConfig returns a configuration parameter with the given ID and true.
// If the parameter was not found it returns false.
func (i *Installation) GetConfig(id string) (*connector.Configuration, bool) {
	for _, c := range i.Configuration {
		if c.ID == id {
			return &c, true
		}
	}
	return nil, false
}

type Instance struct {
	ID             string                       `db:"id" json:"id"`
	InstallationID string                       `db:"installation_id" json:"installationId"`
	Token          connector.InstantiationToken `db:"token" json:"token"`
	ThingID        string                       `db:"thing_id" json:"thingId"`
}
