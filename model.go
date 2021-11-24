package giphy

import "github.com/connctd/connector-go"

type Instance struct {
	ID             string                       `db:"id" json:"id"`
	InstallationID string                       `db:"installation_id" json:"installationId"`
	Token          connector.InstantiationToken `db:"token" json:"token"`
	ThingID        string                       `db:"thing_id" json:"thingId"`
}
