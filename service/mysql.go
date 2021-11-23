package main

import (
	"context"
	"fmt"

	"github.com/connctd/connector-go"
	_ "github.com/db-journey/mysql-driver"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var (
	statementInsertInstallation = `INSERT INTO installations (id, token) VALUES (?, ?)`

	statementInsertInstance = `INSERT INTO instances (id, installation_id, token) VALUES (?, ?, ?)`
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
