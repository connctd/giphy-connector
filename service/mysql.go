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
	statementInsertInstallation = `INSERT INTO installations (id, token) VALUES (?, ?)`

	statementInsertInstance  = `INSERT INTO instances (id, installation_id, token) VALUES (?, ?, ?)`
	statementGetInstanceByID = `SELECT id, token, thing_id FROM instances WHERE id = ?`
	statementGetInstances    = `SELECT id, token, thing_id FROM instances`

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
func (m *DBClient) GetInstances(ctx context.Context) ([]*giphy.Instance, error) {
	var result []*giphy.Instance
	err := m.db.Select(&result, statementGetInstances)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve instance: %w", err)
	}

	return result, nil
}

// AddThingID updates the instance with the thing ID.
func (m *DBClient) AddThingID(ctx context.Context, instanceId string, thingID string) error {
	_, err := m.db.Exec(statementInsertThingId, thingID, instanceId)
	if err != nil {
		return fmt.Errorf("failed to insert thing id: %w", err)
	}

	return nil
}
