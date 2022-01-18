<p align="center">
  <a href="https://docs.connctd.io">
    <img alt="connctd docs" src="./.github/connector-go-banner.png" />
  </a>
</p>

Connctd SDK for creating and publishing Connectors to the connctd platform.

The SDK implements the [connector protocol](https://docs.connctd.io/connector/connector_protocol/) to handle all communication between a connector and the connctd platform as well as the boilerplate needed by most connectors.
It does not assume a specific technology but provides a connector handler implementing the connector protocol, interfaces to use with the handler and a client for the connctd platform.
For most technologies the provided default implementations should be sufficient to quickly implement a connector without the need to implement code specific to the connctd platform.

<!-- TODO Explain structure -->
## Documentation and usage

Documentation for the SDK can be found [here](https://pkg.go.dev/github.com/connctd/connector-go).
See the [connector documentation](https://docs.connctd.io/connector/connectors/) and the [connctd tutorials](https://tutorial.connctd.io/) for more details on connectors, the connector protocol and the [connctd platform](https://connctd.com/).

Most connectors should be able to use the default implementations and embed the default provider to only develop code specific to the connected technology.
The default provider gives access to update and action channels which can be used by the connector to listen to updates and actions sent by the connctd platform. It also implements methods to push updates and action results back to the connctd platform.
Using the default implementations, developing a new connector therefore breaks down to two basic tasks:

  1.  Defining the things to represent the technology at the connctd platform
  2.  Implementing the features specific to the technology

A public connector using this SDK including a detailed tutorial can be found [at Github](https://github.com/connctd/giphy-connector/).

## Structure

The connector SDK is also the default implementation of the connector protocol.
If you want to develop a connector in a different language, you can use it together with the [connector protocol documentation](https://docs.connctd.io/connector/connector_protocol/) as a reference for your own implementation.
The following gives an overview of the structure of the SDK:

```
├── connctd
│   └── things.go             # Domain models for the connctd thing abstraction
├── crypto
│   ├── signing.go            # Signature validation
│   └── signing_test.go
├── db
│   └── default_database.go   # Default database implementation (Sqlite, Mysql, Postgres)
├── provider
│   └── default_provider.go   # Default provider implementation used by default service
├── service
│   └── default_service.go    # Default service implementation used by the connector handler
├── vendor                    # Dependencies
├── client.go                 # Client for the connctd connectorhub
├── client_test.go
├── connhandler.go            # Connector handler implementing endpoints for the connector protocol
├── errors.go                 # Error definitions
├── go.mod
├── go.sum
├── handlers.go               # Signature validation handlers for the connector protocol
├── handlers_test.go
├── LICENSE
├── messages.go               # Definitions of messages used in the connector protocol
├── model.go                  # Models specific to connectors
├── provider.go               # Interface definition used by the default service
├── README.md
└── service.go                # Interface definitions for the service used by the connector handler
```

## Contact

Please use the provided templates for bug reports and feature requests and feel free to contact connctd at info@connctd.com.
