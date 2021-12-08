package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/connctd/connector-go"

	// registers mysql driver at db journeys registry
	_ "github.com/db-journey/mysql-driver"

	// registers postgres driver at db journeys registry
	_ "github.com/db-journey/postgresql-driver"

	// registers sqlite3 driver at db journeys registry
	_ "github.com/mattn/go-sqlite3"

	"github.com/go-logr/logr"
	"github.com/jmoiron/sqlx"
)

type Database interface {
	AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) error
	AddInstallationConfiguration(ctx context.Context, installationId string, config []connector.Configuration) error
	GetInstallations(ctx context.Context) ([]*connector.Installation, error)
	RemoveInstallation(ctx context.Context, installationId string) error

	AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error
	AddInstanceConfiguration(ctx context.Context, instanceId string, config []connector.Configuration) error
	GetInstance(ctx context.Context, instanceId string) (*connector.Instance, error)
	GetInstances(ctx context.Context) ([]*connector.Instance, error)
	GetInstanceByThingId(ctx context.Context, thingId string) (*connector.Instance, error)
	GetInstanceConfiguration(ctx context.Context, instanceId string) ([]connector.Configuration, error)
	RemoveInstance(ctx context.Context, instanceId string) error

	AddThingID(ctx context.Context, instanceID string, thingID string) error
}

type DBOptions struct {
	Driver DBDriverName
	DSN    string
}

var DefaultOptions = &DBOptions{
	Driver: DriverSqlite3,
	DSN:    "default.sqlite3",
}

type DBDriverName string

var (
	DriverMysql      = DBDriverName("mysql")
	DriverPostgresql = DBDriverName("postgres")
	DriverSqlite3    = DBDriverName("sqlite3")
)

var (
	statementInsertInstallation               = `INSERT INTO installations (id, token) VALUES (?, ?)`
	statementInsertInstallationConfig         = `INSERT INTO installation_configuration (installation_id, id, value) VALUES (?, ?, ?)`
	statementGetInstallations                 = `SELECT id FROM installations`
	statementGetConfigurationByInstallationID = `SELECT id, value FROM installation_configuration WHERE installation_id = ?`
	statementRemoveInstallationById           = `DELETE FROM installations WHERE id = ?`

	statementInsertInstance               = `INSERT INTO instances (id, installation_id, token) VALUES (?, ?, ?)`
	statementGetInstanceByID              = `SELECT id, token, installation_id FROM instances WHERE id = ?`
	statementGetInstanceByThingID         = `SELECT id, token, installation_id FROM instances, (SELECT instance_id FROM instance_thing_mapping WHERE thing_id = ? LIMIT 1) mapping WHERE id = instance_id;`
	statementGetInstances                 = `SELECT id, token, installation_id FROM instances`
	statementInsertInstanceConfig         = `INSERT INTO instance_configuration (installation_id, id, value) VALUES (?, ?, ?)`
	statementGetConfigurationByInstanceID = `SELECT id, value FROM instance_configuration WHERE instance_id = ?`
	statementGetThingsByInstanceID        = `SELECT thing_id FROM instance_thing_mapping WHERE instance_id = ?`
	statementRemoveInstanceById           = `DELETE FROM instances WHERE id = ?`

	statementInsertThingId = `INSERT instance_thing_mapping (instance_id, thing_id) VALUES (?, ?)`
)

const (
	statementCreateInstallationTable = `CREATE TABLE installations (
		id CHAR (36) NOT NULL,
		token TEXT NOT NULL,
		UNIQUE(id)
	)`

	statementCreateInstanceTable = `CREATE TABLE instances (
		id CHAR (36) NOT NULL,
		token TEXT NOT NULL,
		installation_id CHAR (36) NOT NULL,
		thing_id CHAR (36) NOT NULL DEFAULT '',
		UNIQUE(id),
		FOREIGN KEY (installation_id)
			REFERENCES installations(id) ON DELETE CASCADE
	)`

	statementCreateInstaceThingMapping = `CREATE TABLE instance_thing_mapping (
		instance_id CHAR (36) NOT NULL,
		thing_id CHAR (36) NOT NULL,
		FOREIGN KEY (instance_id)
			REFERENCES instances(id) ON DELETE CASCADE
	)`

	statementCreateInstallConfigTable = `CREATE TABLE installation_configuration (
		installation_id CHAR (36) NOT NULL,
		id CHAR (36) NOT NULL,
		value VARCHAR (200) NOT NULL,
		FOREIGN KEY (installation_id)
			REFERENCES installations(id) ON DELETE CASCADE
	)`

	statementCreateInstanceConfigTable = `CREATE TABLE instance_configuration (
		instance_id CHAR (36) NOT NULL,
		id CHAR (36) NOT NULL,
		value VARCHAR (200) NOT NULL,
		FOREIGN KEY (instance_id)
			REFERENCES instances(id) ON DELETE CASCADE
	)`
)

var migrationQueries = []string{
	statementCreateInstallationTable,
	statementCreateInstanceTable,
	statementCreateInstaceThingMapping,
	statementCreateInstallConfigTable,
	statementCreateInstanceConfigTable,
}

type DBClient struct {
	DB     *sqlx.DB
	Logger logr.Logger
}

// NewDBClient creates a new mysql client
func NewDBClient(dbOptions *DBOptions, logger logr.Logger) (*DBClient, error) {
	// establish db connection
	db, err := sqlx.Connect(string(dbOptions.Driver), dbOptions.DSN)
	if err != nil {
		return nil, fmt.Errorf("can't connect to db with DSN: %w", err)
	}

	return &DBClient{db, logger}, nil
}

func (m *DBClient) Migrate() error {
	for _, q := range migrationQueries {
		_, err := m.DB.Exec(q)
		if err != nil {
			return fmt.Errorf("failed to migrate db (query: %v) %v", q, err)
		}
	}
	return nil
}

// AddInstallation adds an installation request to the database.
// It assumes that all data is verified beforehand and therefore does not validate anything on it's own.
func (m *DBClient) AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) error {
	_, err := m.DB.Exec(statementInsertInstallation, installationRequest.ID, installationRequest.Token)
	if err != nil {
		return fmt.Errorf("failed to insert installation: %w", err)
	}

	return nil
}

// AddInstallationConfiguration adds all configuration parameters to the database.
func (m *DBClient) AddInstallationConfiguration(ctx context.Context, installationId string, config []connector.Configuration) error {
	for _, c := range config {
		_, err := m.DB.Exec(statementInsertInstallationConfig, installationId, c.ID, c.Value)
		if err != nil {
			return fmt.Errorf("failed to insert installation config: %w", err)
		}
	}

	return nil
}

// GetInstallations returns a list of all existing installations together with their provided configuration parameters.
func (m *DBClient) GetInstallations(ctx context.Context) ([]*connector.Installation, error) {
	var installations []*connector.Installation
	err := m.DB.Select(&installations, statementGetInstallations)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve instance: %w", err)
	}
	for i, installation := range installations {
		var configurations []connector.Configuration
		err := m.DB.Select(&configurations, statementGetConfigurationByInstallationID, installation.ID)
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
	_, err := m.DB.Exec(statementRemoveInstallationById, installationId)
	if err != nil {
		if err == sql.ErrNoRows {
			return connector.ErrorInstallationNotFound
		}
		return fmt.Errorf("failed to remove installation: %w", err)
	}

	return nil
}

// AddInstance adds an instantiation to the database.
func (m *DBClient) AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error {
	_, err := m.DB.Exec(statementInsertInstance, instantiationRequest.ID, instantiationRequest.InstallationID, instantiationRequest.Token)
	if err != nil {
		return fmt.Errorf("failed to insert instance: %w", err)
	}

	return nil
}

// AddInstanceConfiguration adds all configuration parameters to the database.
func (m *DBClient) AddInstanceConfiguration(ctx context.Context, instanceId string, config []connector.Configuration) error {
	for _, c := range config {
		_, err := m.DB.Exec(statementInsertInstanceConfig, instanceId, c.ID, c.Value)
		if err != nil {
			return fmt.Errorf("failed to insert installation config: %w", err)
		}
	}

	return nil
}

// GetInstance returns the instance with the given id.
func (m *DBClient) GetInstance(ctx context.Context, instanceId string) (*connector.Instance, error) {
	var instance connector.Instance
	err := m.DB.Get(&instance, statementGetInstanceByID, instanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve instance: %w", err)
	}

	config, err := m.GetInstanceConfiguration(ctx, instance.ID)
	if err != nil {
		return nil, err
	}
	instance.Configuration = config

	things, err := m.GetThingIDsByInstance(ctx, instance.ID)
	if err != nil {
		return nil, err
	}
	instance.ThingIDs = things

	return &instance, nil
}

// GetInstances returns all instances.
func (m *DBClient) GetInstances(ctx context.Context) ([]*connector.Instance, error) {
	var instances []*connector.Instance
	err := m.DB.Select(&instances, statementGetInstances)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve instance: %w", err)
	}
	for i, instance := range instances {
		config, err := m.GetInstanceConfiguration(ctx, instance.ID)
		if err != nil {
			return nil, err
		}
		instances[i].Configuration = config

		things, err := m.GetThingIDsByInstance(ctx, instance.ID)
		if err != nil {
			return nil, err
		}
		instance.ThingIDs = things
	}
	return instances, nil
}

// GetInstanceByThingId returns the instance with the given thing id.
func (m *DBClient) GetInstanceByThingId(ctx context.Context, thingId string) (*connector.Instance, error) {
	var instance connector.Instance
	err := m.DB.Get(&instance, statementGetInstanceByThingID, thingId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve instance: %w", err)
	}

	config, err := m.GetInstanceConfiguration(ctx, instance.ID)
	if err != nil {
		return nil, err
	}
	instance.Configuration = config

	things, err := m.GetThingIDsByInstance(ctx, instance.ID)
	if err != nil {
		return nil, err
	}
	instance.ThingIDs = things

	return &instance, nil
}

// GetInstanceConfigurations returns all configuration parameters for the given instance id.
// If no parameters where found it return an empty slice.
func (m *DBClient) GetInstanceConfiguration(ctx context.Context, instanceId string) ([]connector.Configuration, error) {
	var configurations []connector.Configuration
	err := m.DB.Select(&configurations, statementGetConfigurationByInstanceID, instanceId)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to retrieve instance configuration")
	}
	return configurations, nil
}

// GetThingIDsByInstance returns all things mapped to the instance with the given id.
func (m *DBClient) GetThingIDsByInstance(ctx context.Context, instanceId string) ([]string, error) {
	var thingIDs []string
	err := m.DB.Select(&thingIDs, statementGetThingsByInstanceID, instanceId)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to retrieve thing ids %v", err)
	}
	return thingIDs, nil
}

// RemoveInstance removes the instance with the given id from the database.
func (m *DBClient) RemoveInstance(ctx context.Context, instanceId string) error {
	_, err := m.DB.Exec(statementRemoveInstanceById, instanceId)
	if err != nil {
		if err == sql.ErrNoRows {
			return connector.ErrorInstanceNotFound
		}
		return fmt.Errorf("failed to remove instance")
	}

	return nil
}

// AddThingID updates the instance with the thing ID.
func (m *DBClient) AddThingID(ctx context.Context, instanceId string, thingId string) error {
	_, err := m.DB.Exec(statementInsertThingId, instanceId, thingId)
	if err != nil {
		return fmt.Errorf("failed to insert thing id: %w", err)
	}

	return nil
}
