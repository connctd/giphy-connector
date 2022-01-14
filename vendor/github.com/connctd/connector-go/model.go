package connector

// Installation represents connector installationa and their configuration.
// We store it in the database to be able to access the configuration and to map instances to specific installations.
// Also we need to store the installation token, to be able to manage the installation with the platform.
// The installation token must be kept secret.
type Installation struct {
	ID            string            `db:"id" json:"id"`
	Token         InstallationToken `db:"token" json:"token"`
	Configuration []Configuration   `json:"configuration"`
}

// GetConfig returns the configuration parameter with the given ID.
// If the parameter was not found it returns false.
func (i *Installation) GetConfig(id string) (*Configuration, bool) {
	for _, c := range i.Configuration {
		if c.ID == id {
			return &c, true
		}
	}
	return nil, false
}

// Configuration is key value pair
type Configuration struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// InstallationToken can be used by a connector installation to propagte e.g. state changes
type InstallationToken string

// Instance represents connector instances.
// We store it in the database to be able to map a instance to a specific thing and installation.
// We also store its token to be able to manage the instance with the platform.
// The instance token must be kept secret.
type Instance struct {
	ID             string             `db:"id" json:"id"`
	InstallationID string             `db:"installation_id" json:"installationId"`
	Token          InstantiationToken `db:"token" json:"token"`
	ThingMapping   []ThingMapping     `json:"things"`
	Configuration  []Configuration    `json:"configuration"`
}

// GetConfig returns the configuration parameter with the given ID.
// If the parameter was not found it returns false.
func (i *Instance) GetConfig(id string) (*Configuration, bool) {
	for _, c := range i.Configuration {
		if c.ID == id {
			return &c, true
		}
	}
	return nil, false
}

// InstantiationToken can be used by a connector instance to send things or updates
type InstantiationToken string

// ThinkMapping represents a mapping of instances to things and external ID.
type ThingMapping struct {
	InstanceID string `db:"instance_id" json:"-"`
	ThingID    string `db:"thing_id" json:"thing_id"`
	ExternalID string `db:"external_id" json:"external_id"`
}
