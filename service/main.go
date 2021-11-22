package main

import (
	"context"
	"encoding/base64"
	ghttp "giphy-connector/http"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
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
		panic("Invalid public key")
	}

	// Create a new instance of our connector
	service := NewService(logrus.WithField("component", "service"))

	// Create a new HTTP handler using the service
	httpHandler := ghttp.MakeHandler(backgroundCtx, publicKey, service)

	// Start the http server using our handler
	http.ListenAndServe(":8080", httpHandler)
}
