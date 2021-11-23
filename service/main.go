package main

import (
	"context"
	"encoding/base64"
	"flag"
	ghttp "giphy-connector/http"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	dsn := flag.String("mysql.dsn", "", "DSN in order to connect with db")

	flag.Parse()

	backgroundCtx := context.Background()

	// Request from the connctd platform are signed using the connector publication key
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

	// Create a new database client
	dbClient, err := NewDBClient(*dsn, logrus.WithField("component", "database"))
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// Create a new instance of our connector
	service := NewService(dbClient, logrus.WithField("component", "service"))

	// Create a new HTTP handler using the service
	httpHandler := ghttp.MakeHandler(backgroundCtx, publicKey, service)

	logrus.Info("Initialized connector, start callback handler")
	// Start the http server using our handler
	http.ListenAndServe(":8080", httpHandler)
}
