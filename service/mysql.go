package main

import (
	"context"
	"fmt"
	"giphy-connector"

	"github.com/connctd/connector-go"
	_ "github.com/db-journey/mysql-driver"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var (
	statementInsertInstallation               = `INSERT INTO installations (id, token) VALUES (?, ?)`
	statementInsertInstallationConfig         = `INSERT INTO installation_configuration (installation_id, id, value) VALUES (?, ?, ?)`
	statementGetInstallations                 = `SELECT id FROM installations`
	statementGetConfigurationByInstallationID = `SELECT id, value FROM installation_configuration WHERE installation_id = ?`
	statementRemoveInstallationById           = `DELETE FROM installtations WHERE id = ?`

	statementInsertInstance       = `INSERT INTO instances (id, installation_id, token) VALUES (?, ?, ?)`
	statementGetInstanceByID      = `SELECT id, token, installation_id, thing_id FROM instances WHERE id = ?`
	statementGetInstanceByThingID = `SELECT id, token, installation_id, thing_id FROM instances WHERE thing_id = ?`
	statementGetInstances         = `SELECT id, token, installation_id, thing_id FROM instances`
	statementRemoveInstanceById   = `DELETE FROM instances WHERE id = ?`

	statementInsertThingId = `UPDATE instances SET thing_id = ? WHERE id = ?`
)

type DBClient struct {
	db     *sqlx.DB
	logger logrus.FieldLogger
}

// NewDBClient creates a new mysql client
func NewDBClient(dsn string, logger logrus.FieldLogger) (*DBClient, error) {
	// establish db connection
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("can't connect to mysql db with DSN: %w", err)
	}

	return &DBClient{db, logger}, nil
}

// AddInstallation adds an installation request to the database.
// It assumes that all data is verified beforehand and therefore does not validate anything on it's own.
func (m *DBClient) AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) error {
	_, err := m.db.Exec(statementInsertInstallation, installationRequest.ID, installationRequest.Token)
	if err != nil {
		return fmt.Errorf("failed to insert installation: %w", err)
	}

	return nil
}

// AddInstallationConfiguration adds all configuration parameters to the database.
func (m *DBClient) AddInstallationConfiguration(ctx context.Context, installationId string, config []connector.Configuration) error {
	for _, c := range config {
		_, err := m.db.Exec(statementInsertInstallationConfig, installationId, c.ID, c.Value)
		if err != nil {
			return fmt.Errorf("failed to insert installation config: %w", err)
		}
	}

	return nil
}

// GetInstallations returns a list of all existing installations together with their provided configuration parameters.
// It is used to register all installations with the Giphy provider at startup.
func (m *DBClient) GetInstallations(ctx context.Context) ([]*giphy.Installation, error) {
	var installations []*giphy.Installation
	err := m.db.Select(&installations, statementGetInstallations)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve instance: %w", err)
	}
	for i, installation := range installations {
		var configurations []connector.Configuration
		err := m.db.Select(&configurations, statementGetConfigurationByInstallationID, installation.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve instance: %w", err)
		}
		installations[i].Configuration = configurations
	}
	return installations, nil
}

// RemoveInstance removes the instance with the given id from the database.
// This will also remove instances belonging to this installation, as well as the configuration parameters.
// Removal of config parameters and instances is implemented via cascading foreign keys in the database.
// If your database does not support cascading foreign keys, you should delete them manually.
func (m *DBClient) RemoveInstallation(ctx context.Context, installationId string) error {
	_, err := m.db.Exec(statementRemoveInstallationById, installationId)
	if err != nil {
		return fmt.Errorf("failed to remove instance")
	}

	return nil
}

// AddInstance adds an instantiation to the database.
func (m *DBClient) AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error {
	_, err := m.db.Exec(statementInsertInstance, instantiationRequest.ID, instantiationRequest.InstallationID, instantiationRequest.Token)
	if err != nil {
		return fmt.Errorf("failed to insert instance: %w", err)
	}

	return nil
}

// GetInstance returns the instance with the given id.
func (m *DBClient) GetInstance(ctx context.Context, instanceId string) (*giphy.Instance, error) {
	var result giphy.Instance
	err := m.db.Get(&result, statementGetInstanceByID, instanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve instance: %w", err)
	}

	return &result, nil
}

// GetInstances returns all instances.
// It is used to register all instances with the Giphy provider at startup.
func (m *DBClient) GetInstances(ctx context.Context) ([]*giphy.Instance, error) {
	var result []*giphy.Instance
	err := m.db.Select(&result, statementGetInstances)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve instance: %w", err)
	}

	return result, nil
}

// GetInstanceByThingId return the instance with the given thing id.
func (m *DBClient) GetInstanceByThingId(ctx context.Context, thingId string) (*giphy.Instance, error) {
	var result giphy.Instance
	err := m.db.Get(&result, statementGetInstanceByThingID, thingId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve instance: %w", err)
	}

	return &result, nil
}

// RemoveInstance removes the instance with the given id from the database.
func (m *DBClient) RemoveInstance(ctx context.Context, instanceId string) error {
	_, err := m.db.Exec(statementRemoveInstanceById, instanceId)
	if err != nil {
		return fmt.Errorf("failed to remove instance")
	}

	return nil
}

// AddThingID updates the instance with the thing ID.
func (m *DBClient) AddThingID(ctx context.Context, instanceId string, thingID string) error {
	_, err := m.db.Exec(statementInsertThingId, thingID, instanceId)
	if err != nil {
		return fmt.Errorf("failed to insert thing id: %w", err)
	}

	return nil
}
