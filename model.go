package giphy

import "github.com/connctd/connector-go"

const (
	RandomComponentId       = "random"
	RandomPropertyId        = "value"
	SearchComponentId       = "search"
	SearchPropertyId        = "value"
	SearchActionId          = "search"
	SearchActionParameterId = "keyword"
)

// Installation represents connector installationa and their configuration.
// We store it in the database to be able to access the configuration and to map instances to specific installations.
// Also we need to store the installation token, to be able to manage the installation with the platform.
// The installation token must be kept secret.
type Installation struct {
	ID            string                      `db:"id" json:"id"`
	Token         connector.InstallationToken `db:"token" json:"token"`
	Configuration []connector.Configuration   `json:"configuration"`
}

// GetConfig returns a configuration parameter with the given ID.
// If the parameter was not found it returns false.
func (i *Installation) GetConfig(id string) (*connector.Configuration, bool) {
	for _, c := range i.Configuration {
		if c.ID == id {
			return &c, true
		}
	}
	return nil, false
}

// Instance represents connector instances.
// We store it in the database to be able to map a instance to a specific thing and installation.
// We also store its token to be able to manage the instance with the platform.
// The instance token must be kept secret.
type Instance struct {
	ID             string                       `db:"id" json:"id"`
	InstallationID string                       `db:"installation_id" json:"installationId"`
	Token          connector.InstantiationToken `db:"token" json:"token"`
	ThingID        string                       `db:"thing_id" json:"thingId"`
}
