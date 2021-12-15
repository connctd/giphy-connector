package main

import (
	"context"
	"encoding/base64"
	"flag"
	"net/http"
	"os"

	"github.com/connctd/connector-go"
	"github.com/connctd/connector-go/db"
	"github.com/connctd/connector-go/service"
)

func main() {
	migrate := flag.Bool("migrate", false, "")

	flag.Parse()

	// Requests from the connctd platform are signed using the connector publication key
	// To verify the signature, we need the coresponding public key, which we retrieve during connector publication
	key := os.Getenv("GIPHY_CONNECTOR_PUBLIC_KEY")
	if key == "" {
		panic("GIPHY_CONNECTOR_PUBLIC_KEY environment variable not set")
	}
	// To use the retrieved public key, we need to decode it first
	publicKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		panic("Invalid public key: " + err.Error())
	}

	// Create the Giphy provider
	giphyProvider := NewGiphyProvider()

	// Create a new database client
	// Uncomment the next lines to use a mysql database
	// dbOptions := &db.DBOptions{
	// 	Driver: db.DriverMysql,
	// 	DSN:    "root@tcp(localhost)/giphy_connector?parseTime=true",
	// }
	// dbClient, err := db.NewDBClient(dbOptions, connector.DefaultLogger)

	// Uses a Sqlite3 database by default
	dbClient, err := db.NewDBClient(db.DefaultOptions, connector.DefaultLogger)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// Migrate the database if the flag was set
	// The database should only be migrated once
	if *migrate {
		if err := dbClient.Migrate(); err != nil {
			panic("Failed to migrate database " + err.Error())
		}
	}

	// Create a new client for the connctd API
	connctdClient, err := connector.NewClient(nil, connector.DefaultLogger)
	if err != nil {
		panic("Failed to create connctd client: " + err.Error())
	}

	// Create a new instance of our connector
	service, err := service.NewConnectorService(dbClient, connctdClient, giphyProvider, thingTemplate, connector.DefaultLogger)
	if err != nil {
		panic("Failed to create connector service: " + err.Error())
	}

	// Start the event handler listening to action and property update events
	ctx := context.Background()
	service.EventHandler(ctx)

	// Create a new HTTP handler using the service
	httpHandler := connector.NewConnectorHandler(nil, service, publicKey)

	connector.DefaultLogger.Info("start giphy provider")
	// Start Giphy provider
	giphyProvider.Run()

	connector.DefaultLogger.Info("start callback handler")
	// Start the http server using our handler
	err = http.ListenAndServe(":8080", httpHandler)
	if err != nil {
		connector.DefaultLogger.Error(err, "failed to start handler")
	}
}
