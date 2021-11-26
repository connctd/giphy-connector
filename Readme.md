## Connector Tutorial

In this tutorial we are going to implement a connector to connect [Giphy](https://giphy.com/) to our platform.
Source code for this tutorial can be found at [Github](https://github.com/connctd/giphy-connector).

### Requirements

The connector connects Giphy to the connctd platform.
Therefore you need an account for the [connctd Developer Center](https://devcenter.connctd.io/).

We are using the Giphy API which requires an account with Giphy and a Giphy API key.
See the [Giphy documentation](https://developers.giphy.com/docs/api#quick-start-guide) on how to acquire them.

We will implement the connector in [Go](https://golang.org/) and assume basic knowledge of the language.
However, the required steps to implement a connector are more or less the same in every language.
Using Go allows us to use the [connector-go](https://github.com/connctd/connector-go) library provided by connctd which will simplify some things.

The connector must be reachable from the connctd platform.
For this tutorial we use [ngrok](https://ngrok.com/) to create public URLs for our local connector.
While this is not a production ready solution, it gives us a simple development environment and lets us quickly deploy changes to our connector.
In a production environment, you can use a hosting provider of your choice.

### Connector

Conceptionally a connector consists of two parts:

1. A HTTP handler, that provides endpoints for the connector protocol
2. A service that implements the technology we want to connect to the connctd platform and communicates to the connctd platform via API provided by the platform

In this tutorial our focus is on the first part as the second part is specific to the technology you want to connect to the platform.
Nevertheless we use the second part to show the API provided by the connctd platform, which is an important part of connector development.

For more information on the connectors and an in-depth description of the connector protocol, take a look at our [documentation](https://docs.connctd.io/connector/connectors/).
For a better understanding on how connectors work we recommend to read at least the introduction to connectors and the connector protocol.

TODO: Overview of endpoints to implement: installation and instantiations (POST / DELETE), action requests

### First steps

Create a new directory (`giphy-connector`) and initialize a new Go module in it with `go mod init`.
This will create a **go.mod** which is used to specify our module and dependencies.
We will add some directories and files to structure our connectors:

* **http/**
  * **handler.go**

    Contains the HTTP handler providing the callbacks for the connector protocol.

  * **http.go**

    Used to setup the HTTP handler and routes.

* **service/**

  * **main.go**
    * **main.go** 
  * **main.go**
    * **main.go** 
  * **main.go**

    The entry point to our connector.
    Starts the HTTP handler and the connector.

  * **protocol.go**

    Implements the connector protocol functions used by the HTTP handler.

* **vendor/**

    Contains the dependencies of our connector.

* **service.go**

    Defines a service interface implemented by our connector.
    Can be used in the HTTP handler to handle the connector protocol.

If you do not want to start from scratch you can also check out the code repository at tag [**step1**](https://github.com/connctd/giphy-connector/tree/step1) for a basic code template containing the structure explained above.
The tag also contains some code to setup the callback endpoints defined in the connector protocol.

In **http.go** we set up the routes for the callbacks, as well as the signature verification.
Note that some routes, namely the routes for installation and instance removal.
We will add them later on.

**handler.go** defines some basic handler functions for the callbacks.
At this point the handler will verify the required request body but do nothing else.

### Implementing Callback Handlers

In the next step we are going to implement the handler functions to install and instantiate connectors.
We will concentrate on the install handler, since both handler are very similar to implement.

Three steps are needed to implement the handler:

1. Implement the service in **protocol.go**.
   In our first implementation the service will just store the request data in the database.
   We could also implement other things here, like validation of the configuration parameters.
2. Implement the HTTP handler in **handler.go**.
   They will delegate installations and instantiations to the injected service and respond to the connctd platform.
3. Implement a database client to be able to persist all relevant data.

The service needs to store data related to the installations and instantiations in a database.
We will use [MySQL](https://www.mysql.com/) but connectors in general do not require a particular database.
An easy way to start a local MySQL instance is to use Docker.

The following command will start a MySQL in a Docker container that is reachable from the host on port 3306 and also create a database called **giphy_connector**.

`docker run --name giphy-mysql -p 3306:3306 -e MYSQL_ALLOW_EMPTY_PASSWORD=true -e MYSQL_DATABASE=giphy_connector -d mysql:5.7`

For more information about Docker and running MySQL in a Docker container, please refer to the [official documentation](https://hub.docker.com/_/mysql/).

### Service Implementation

We start by implementing the service for installations.
We know that we need to store the installation in the database, so we create a database interface with one method and inject the interface into our service.

**service.go**
```golang
type Database interface {
    AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) error
}
```

**protocol.go**
```golang
type GiphyConnector struct {
    logger logrus.FieldLogger
    db     giphy.Database
}

// NewService returns an new instance of the Giphy connector
func NewService(dbClient giphy.Database, logger logrus.FieldLogger) giphy.Service {
    return &GiphyConnector{
        logger,
        dbClient,
    }
}
```

As mentioned above, the service just stores the installation request in the database.
We do this by calling our database interface method.

**protocol.go**
```golang
func (g *GiphyConnector) AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) (*connector.InstallationResponse, error) {
    logrus.WithField("installationRequest", installationRequest).Infoln("Received an installation request")

    if err := g.db.AddInstallation(ctx, installationRequest); err != nil {
        g.logger.WithError(err).Errorln("Failed to add installation")
        return nil, err
    }

    return nil, nil
}
```

Next we will use this service to implement our HTTP handler.
The handler will call the service with the request and respond to the connctd platform with an appropriate return code.

**handler.go**

```golang
func handleInstallation(service giphy.Service) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var req connector.InstallationRequest
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        if err = json.Unmarshal(body, req); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        response, err := service.AddInstallation(r.Context(), req)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        // Installation was successfull so far but we require further steps
        if response != nil {
            // We add the further steps and details to the response
            // We set the status code to accepted
            // This signals the connctd platform, that the installation is ongoing
            w.WriteHeader(http.StatusAccepted)
            // As a good citizen we also set an appropriate content type
            w.Header().Add("Content-Type", "application/json")

            b, err := json.Marshal(response)
            if err != nil {
                // Abort if we cannot marshal the response
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte("{\"err\":\"Failed to marshal error\"}"))
            } else {
                w.Write(b)
            }
            return
        }

        // Installation was successful and we do not need any further steps
        // No payload is added to the response
        w.WriteHeader(http.StatusCreated)
    })
}
```

Next we have to implement the database client.
We create a new file **mysql.go** in the **service** directory and add the following code, that lets us create a new database client and use it in our service.

**mysql.go**

```golang
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
```

We will add a command line flag to our **main.go** using the build in **flag** package.
This lets us choose the DSN for our database at startup.

**main.go**

```golang
func main() {
    dsn := flag.String("mysql.dsn", "", "DSN in order to connect with db")

    flag.Parse()

    [...]
    
    // Create a new database client
    dbClient, err := NewDBClient(*dsn, logrus.WithField("component", "database"))
    if err != nil {
        panic("Failed to connect to database: " + err.Error())
    }

    // Create a new instance of our connector
    service := NewService(dbClient, logrus.WithField("component", "service"))

    // Create a new HTTP handler using the service
    httpHandler := ghttp.MakeHandler(backgroundCtx, publicKey, service)

    // Start the http server using our handler
    http.ListenAndServe(":8080", httpHandler)
}
```

To be able to use our database client in our service, it has to implement the database interface.
For now, the connector stores only the installation ID and the corresponding token.
We will create a migration that creates a installations table with two columns and store it in **service/migration/0001_init.up.sql**.
From our working directory, we can then migrate our database with the following command (assuming you have started MySQL using the command explained above).

`docker exec -i giphy-mysql mysql giphy_connector < service/migrations/0001_init.up.sql`

**0001_init.up.sql**

```sql
CREATE TABLE installations (
    id CHAR (36) NOT NULL,
    token TEXT NOT NULL,
    UNIQUE(id)
);
```

Now we can implement the database client:

**mysql.go**

```golang
var (
    statementInsertInstallation = `INSERT INTO installations (id, token) VALUES (?, ?)`
)
func (m *DBClient) AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) error {
    _, err := m.db.Exec(statementInsertInstallation, installationRequest.ID, installationRequest.Token)
    if err != nil {
        return fmt.Errorf("failed to insert installation: %w", err)
    }

    return nil
}
```

With the installation endpoint ready, we could in theory publish our connector to the connctd platform and successfully install it.
Nevertheless, to actually create Things and update their status, we need *instantiations* of the installed connector.
So before publishing our connector, we will implement the instantiation endpoint.
Since it is very similar to the installation endpoint, we will not explain it here in detail.

You can find a version of the connector with both endpoints implemented at tag [**step2**](https://github.com/connctd/giphy-connector/tree/step2).
We have also added a **Makefile** which lets you build the project with `make build` and a script that lets you run the connector with `./runs.sh`.

With that in place it's time for a first test run of our connector.
Start the connector with the provided run script and then start ngrok with `ngrok http 8080`.
ngrok should give you two URLs and show some logs.
We use the URL using HTTPS, since HTTP is not supported by the connctd platform.
To verify that everything works, we can send a POST request to the URL via curl:

`curl --location --request POST 'https://-ngrokId-.ngrok.io/callbacks/installations'`

ngrok should print a log for the request and the connector should have responded with status code **400** and some JSON:

```json
{
    "error": "MISSING_HEADER",
    "description": "Signable payload can not be generated since a relevant header is missing",
    "status": 400
}
```

This is our SignatureValidationHandler in action, which refuses the request since we didn't provide a valid signature.

### Connector Publication

We can now publish our connector to the connctd platform.
Note however that we may have to republish the connector, when the ngrok URL changes.
To publish the connector, login to the [Developer Center](https://devcenter.connctd.io/) and create a new App, if you haven't already.
Make sure to safely store the client id and client secret, we will need it later on and it can not be restored.
Next go to the [Connector Store](https://devcenter.connctd.io/connectors) and click "Create Connector".
TODO: Explain first step?
In the second step, provide the callback URLs:

installationCallbackUrl: https://-ngrokId-.ngrok.io/callbacks/installations

instanceCallbackUrl: https://-ngrokId-.ngrok.io/callbacks/instantiations

actionCallbackUrl: https://-ngrokId-.ngrok.io/callbacks/actions

Complete the publication and make sure to save the public key which is shown after the publication.
Before we install the connector with the app, we have to provide the public key to the connector.
To do this take a look at **run.sh** and replace the key exported for GIPHY_CONNECTOR_PUBLIC_KEY with the new key.
After this, restart the connector (do not restart ngrok).

You can now install the connector.
Select the new App in the Developer Center sidebar then go to the Connector Store.
Click "Install from Connector Store".
You should see the connector listed under your private connectors.
Click on it and click the install button.
The installation should be successful and the connector should log the installation request.

Creating an instance is not possible with the Developer Center, since it is usually done programmatically by the App using a connector.
We can however use the connctd GraphQL API to create an instance manually.
For this you need the installation ID, which you can find in the [Developer Center](https://devcenter.connctd.io/connectors).
Click on your installation and the ID is shown.
Also you need an authorization token which you can acquire by TODO: ClientCredentialFlow.

To instantiate the connector send the following request, replacing your auth token and installation ID.
If you get an error saying that the instance was already created, set the `X-External-Subject_ID` to something else.

```http
curl --location --request POST 'https://api.connctd.io/api/v1/query' \
--header 'Authorization: Bearer --yourAuthToke--' \
--header 'Content-Type: application/json' \
--data-raw '{"query":"mutation InstantiateConnector {\n    instantiateConnector(installationId: \"--installtionID--\", configuration: []\n    ) {\n        id\n        installation {\n            id\n        }\n        state\n        stateDetails\n    }\n}","variables":{}}'
```

TODO Clean up query / delete \n
(TODO Should we provide postman requests as an alternative?)

### Thing Abstraction

Now we have an installation and an instance of our connector and can start adding Things to the instance.
A Thing is an abstract representation of a specific technology.
It could for example represent a sensor that will periodically update it's value or a smart lock that can be opened and closed via the connctd platform.
In our case it will act as an *virtual sensor* that will periodically update its value to a random Gif from the Giphy platform.
Later on we will add an action to the Thing but for now, we want a minimal connector that can interact with the connctd platform.

If you remember the [introduction](#connector), we will now start to implement the second part of the connector that connects Giphy with the connctd platform.
We divide this work in two steps: first we represent Giphy as Thing and create such a Thing for each instance of our connector.
Second we implement a client for a subset of the Giphy API to periodically update the created Things.

In general, a Thing has the following form:

```golang
type Thing struct {
    ID              string          
    Name            string          
    Manufacturer    string          
    DisplayType     string          
    MainComponentID string          
    Status          StatusType      
    Components      []Component     
    Attributes      []ThingAttribute
}


type Component struct {
    ID            string    
    Name          string    
    ComponentType string    
    Capabilities  []string  
    Properties    []Property
    Actions       []Action  
}

type Property struct {
    ID           string   
    Name         string   
    Value        string   
    Unit         string   
    Type         ValueType
    LastUpdate   time.Time
    PropertyType string   
}
```

See [connctd/restapi-go](https://github.com/connctd/restapi-go/blob/main/things.go) for more details.
The important part here is that a Thing has a set of components, which by themselves have a set of properties and actions.
The properties will represent all values that the Thing can produce.
Actions represent actions that a Thing can invoke.
For now we concentrate on the properties.
For each new instance we will create a Thing with a property with the type string so it can hold a URL to the latest random Gif from Giphy.
Since the Thing is managed by the connctd platform, we only have to store its ID to be able to match a thing to its instance.
Luckily, the [Connector SDK](https://github.com/connctd/connector-go) already implements Things and a client for the connctd platform that lets us manage them.

We will register a new Thing with the platform and save its ID together with the instance whenever a new instance is created.
To do this, we change **protocol.go** to create a thing whenever the instantiation handler is called.
Since we only have one Thing per instance, we can just save the Thing ID in the same table.
To make things easy, we use the connctd client from the SDK to handle the details.

**protocol.go**

```golang
type GiphyConnector struct {
    logger        logrus.FieldLogger
    db            giphy.Database
    connctdClient connector.Client
}

// NewService returns a new instance of the Giphy connector
func NewService(dbClient giphy.Database, connctdClient connector.Client, logger logrus.FieldLogger) giphy.Service {
    return &GiphyConnector{
        logger,
        dbClient,
        connctdClient,
    }
}

func (g *GiphyConnector) AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error {
    logrus.WithField("instantiationRequest", instantiationRequest).Infoln("Received an instantiation request")

    if err := g.db.AddInstance(ctx, instantiationRequest); err != nil {
        g.logger.WithError(err).Errorln("Failed to add instance")
        return err
    }

    return g.CreateThing(ctx, instantiationRequest.ID)
}
```

We add a new file **service/things.go** that will contain our thing related code.
Here we build the Things and register them with the platform.
We know that we will want to update the property of the Things, so we will also implement the update here.

**things.go**

```golang
// buildThing returns a thing that can be registered with the connctd platform.
// Note that the thing ID is generated by connctd and returned when the thing is created.
// The connctd platform will store all information regarding the thing.
// The connector therefore should only store its ID.
func buildThing() restapi.Thing {
    return restapi.Thing{
        Name:            "Giphy",
        Manufacturer:    "IoT connctd GmbH",
        DisplayType:     "core.SENSOR",
        MainComponentID: randomComponentId,
        Status:          "AVAILABLE",
        Attributes:      []restapi.ThingAttribute{},
        Components: []restapi.Component{
            {
                ID:            randomComponentId,
                Name:          "Giphy random component",
                ComponentType: "core.Sensor",
                Capabilities: []string{
                    "core.MEASURE",
                },
                Properties: []restapi.Property{
                    {
                        ID:    randomPropertyId,
                        Name:  "Giphy random component",
                        Value: "",
                        Type:  restapi.ValueTypeString,
                    },
                },
                Actions: []restapi.Action{},
            },
        },
    }
}

// CreateThing can be called by the connector to register a new thing for the given instance.
// It retrieves the instance token from the database and uses the token to create a new thing via the connctd API client.
// The new thing ID is then stored in the database referencing the instance id.
func (g *GiphyConnector) CreateThing(ctx context.Context, instanceId string) error {
    instance, err := g.db.GetInstance(ctx, instanceId)
    if err != nil {
        g.logger.WithField("instanceId", instanceId).WithError(err).Error("failed to retrieve instance from database")
        return err
    }

    thing := buildThing()
    createdThing, err := g.connctdClient.CreateThing(ctx, instance.Token, thing)
    if err != nil {
        g.logger.WithField("thing", thing).WithError(err).Error("failed to register new Thing")
        return err
    }

    err = g.db.AddThingID(ctx, instanceId, createdThing.ID)
    if err != nil {
        g.logger.WithField("thing", thing).WithError(err).Error("failed to insert new Thing into database")
        return err
    }

    g.logger.WithField("thing", createdThing).Info("Created new thing")

    return nil
}

// UpdateTrending can be called by the connector to update the trending property of an thing belonging to an instance.
func (g *GiphyConnector) UpdateTrending(ctx context.Context, instanceId string, value string) error {
    instance, err := g.db.GetInstance(ctx, instanceId)
    if err != nil {
        g.logger.WithField("instanceId", instanceId).WithError(err).Error("failed to retrieve instance")
        return err
    }
    if instance.ThingID == "" {
        g.logger.WithField("instanceId", instanceId).Error("Thing id not set")
        return errors.New("thing id not set")
    }

    timestamp := time.Now()

    err = g.connctdClient.UpdateThingPropertyValue(ctx, instance.Token, instance.ThingID, trendingComponentId, trendingPropertyId, value, timestamp)

    return err
}
```

To be able to create Things and store the IDs in the database, we also extend and implement our database interface.
We also create **model.go** which contains the Go representation of our database model.

**service.go**

```golang
type Connector interface {
    CreateThing(ctx context.Context, instanceId string) error
    UpdateProperty(ctx context.Context, thingId string, value string) error
}

type Database interface {
    AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) error

    AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error
    GetInstance(ctx context.Context, instanceId string) (*Instance, error)

    AddThingID(ctx context.Context, instanceID string, thingID string) error
}
```

**model.go**

```golang
type Instance struct {
    ID             string                       `db:"id" json:"id"`
    InstallationID string                       `db:"installation_id" json:"installationId"`
    Token          connector.InstantiationToken `db:"token" json:"token"`
    ThingID        string                       `db:"thing_id" json:"thingId"`
}
```

With this in place, you should now test the connector and create some new instances.
For every instance you create, a new thing should be created but no property updates should happen yet.
For the complete code see tag [**step3**](https://github.com/connctd/giphy-connector/tree/step3).

### Giphy Provider

We now have installations, instances and things registered at the connctd platform but nothing in place to update our things.
If you query your thing (see [documentation](https://docs.connctd.io/graphql/things/) or use the [Developer Center](https://devcenter.connctd.io/things)) the value of the property should be an empty string.
To update the property we will implement a Giphy provider, that will periodically call the Giphy API to get a new random Gif and send an event for every registered instance to the connector.
We will use an [API client](https://github.com/peterhellberg/giphy) for the Giphy API to keep things simple.
Note that this implementation will use the same API key for all installations and instances.
This is usually not what we want and we will hit rate limiting pretty fast with this approach.
Depending on the use case we should instead provide an API key during installation or instantiation using configuration parameters.
We will implement configuration parameters later on.

The provider will implement the following interface, which lets the connector register new instances and listen to the update event channel.

**giphyhandler.go**

```golang
type Provider interface {
    UpdateChannel() <-chan GiphyUpdate
    RegisterInstances(instanceId ...string) error
}

type GiphyUpdate struct {
    InstanceId string
    Value      string
}
```

The connector can then listen to the event channel and update the properties accordingly.
For now, the events will only contain the instance ID and the new value, which is sufficient to update the correct property.
Take a look at **giphy/giphy.go**.
It implements the interface and also the periodic updates by running an infinite loop that we can start in a goroutine.
At the start of each run of the loop it will add the newly registered instances and retrieve a new random Gif,
It will then send an update event for each instance and sleep until the next update.

**giphy.go**

```golang
func (h *Handler) UpdateChannel() <-chan connector.GiphyUpdate {
    return h.updateChannel
}

func (h *Handler) RegisterInstances(instanceIds ...string) error {
    h.newInstances = append(h.newInstances, instanceIds...)
    return nil
}

func (h *Handler) Run() {
    for {
        h.addNewInstances()
        randomGif, err := h.getRandomGif()
        for _, instanceId := range h.instances {
            if err != nil {
                continue
            }
            h.updateChannel <- connector.GiphyUpdate{
                InstanceId: instanceId,
                Value:      randomGif,
            }
        }
        time.Sleep(1 * time.Minute)
    }
}
```

With this in place, we will have to register our instances, start the provider and listen to the update events.

We will register instances in two places: whenever a new instance is created and at start up to register the existing instances.

**protocol.go**

```golang

func (g *GiphyConnector) init() {
    instances, err := g.db.GetInstances(context.Background())
    if err != nil {
        g.logger.WithError(err).Error("Failed to retrieve instances from db")
        return
    }

    instanceIds := make([]string, len(instances))
    for i := range instances {
        instanceIds[i] = instances[i].ID
    }
    g.giphyProvider.RegisterInstances(instanceIds...)

    go g.giphyEventHandler(context.Background())
}

// AddInstantiation is called by the HTTP handler when it retrieved an instantiation request
func (g *GiphyConnector) AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error {
    logrus.WithField("instantiationRequest", instantiationRequest).Infoln("Received an instantiation request")

    if err := g.db.AddInstance(ctx, instantiationRequest); err != nil {
        g.logger.WithError(err).Errorln("Failed to add instance")
        return err
    }

    if err := g.CreateThing(ctx, instantiationRequest.ID); err != nil {
        g.logger.WithError(err).Errorln("Failed to create new thing")
        return err
    }

    g.giphyProvider.RegisterInstances(instantiationRequest.ID)

    return nil
}
```

We will listen to the update events and use our UpdateProperty(..) method to update the properties.
The event listener will also run in a goroutine.

**protocol.go**

```golang
// giphyEventHandler handles events coming from the giphy provider
func (g *GiphyConnector) giphyEventHandler(ctx context.Context) {
    // wait for Giphy events
    go func() {
        for update := range g.giphyProvider.UpdateChannel() {
            g.logger.WithField("value", update.Value).Infoln("Received update from Giphy provider")
            g.UpdateProperty(ctx, update.InstanceId, update.Value)
        }
    }()
}
```

In **main.go** we will start the Giphy provider and configure the required API Key.

**main.go**

```golang
func main() {
    [...]
    
        // We need a Giphy API key to use their API
    giphyAPIKey := os.Getenv("GIPHY_API_KEY")
    if giphyAPIKey == "" {
        panic("GIPHY_API_KEY environment variable not set")
    }
    giphyProvider := gProvider.New(giphyAPIKey)

    // Create a new instance of our connector
    service := NewService(dbClient, connctdClient, giphyProvider, logrus.WithField("component", "service"))

    // Start Giphy provider
    go giphyProvider.Run()
}
```

We have also added a new method to our database to be able to retrieve all existing instances.

**mysql.go**

```golang
// GetInstances returns all instances.
func (m *DBClient) GetInstances(ctx context.Context) ([]*giphy.Instance, error) {
    var result []*giphy.Instance
    err := m.db.Select(&result, statementGetInstances)
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve instance: %w", err)
    }

    return result, nil
}
```

If we now run the connector (don't forget to provide your Gihpy API key in **run.sh**), the property of your registered Thing should now update once a minute.
You can query the history of the property using the [GraphQL API](https://docs.connctd.io/graphql/history/#resolve).
For the complete code see tag [**step4**](https://github.com/connctd/giphy-connector/tree/step4).

### Configuration Parameters

We mentioned [above](#giphy-provider) that we do not want to use the same API key for all installations.
We will now solve this problem by implementing an installation level configuration parameter.
Every App developer who installs and uses the connector will then have to provide his own API key.
If instead every end user should provide a key, we would implement an instantiation configuration parameter.

In order to let the connctd platform know about the configuration parameter, we have to provide it during connector publication.
Lets publish a new connector to add the parameter.
Follow the instructions given [above](#connector-publication) and instead of skipping step 3, provide add a new entry to the installation configuration with the ID ***giphy_api_key***, a value type of ***STRING(Text)*** and required set to true.
Do not forget to click the *Add entry* icon (+) after providing the configuration.
We should pick something meaningful for the ID and name, because it will be used in our connector and shown during installation.
Now start a installation for the new connector.
The installation wizard should ask you to provide an API key.
Before finishing the installation, lets modify our connector to store the provided key.

We create a new database table **installation_configuration** and add the configuration parameters to the database whenever the installation handler is called.

**service/migrations/0002_installations_config.up.sql**

```sql
CREATE TABLE installation_configuration (
    installation_id CHAR (36) NOT NULL,
    id CHAR (36) NOT NULL,
    value VARCHAR (200) NOT NULL,
    FOREIGN KEY (installation_id)
        REFERENCES installations(id)
);
```

Again, you can apply the migration with `docker exec -i giphy-mysql mysql giphy_connector < service/migrations/0002_installation_config.up.sql`.

**protocol.go**

```golang
// AddInstallation is called by the HTTP handler when it retrieved an installation request
func (g *GiphyConnector) AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) (*connector.InstallationResponse, error) {
    logrus.WithField("installationRequest", installationRequest).Infoln("Received an installation request")

    if err := g.db.AddInstallation(ctx, installationRequest); err != nil {
        g.logger.WithError(err).Errorln("Failed to add installation")
        return nil, err
    }

    if err := g.db.AddInstallationConfiguration(ctx, installationRequest.ID, installationRequest.Configuration); err != nil {
        g.logger.WithError(err).WithField("config", installationRequest.Configuration).Errorln("Failed to add installation configuration")
        return nil, err
    }

    return nil, nil
}
```

**mysql.go**

```golang
var (
	statementInsertInstallationConfig = `INSERT INTO installation_configuration (installation_id, id, value) VALUES (?, ?, ?)`
)

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
```

Don't forget to add the new database method to the interface:

**service.go**

```golang
type Database interface {
    AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) error
    AddInstallationConfiguration(ctx context.Context, installationId string, config []connector.Configuration) error

    AddInstance(ctx context.Context, instantiationRequest connector.InstantiationRequest) error
    GetInstance(ctx context.Context, instanceId string) (*Instance, error)
    GetInstances(ctx context.Context) ([]*Instance, error)

    AddThingID(ctx context.Context, instanceID string, thingID string) error
}
```

This is also a good moment to delete all existing installations and instances from our database.
After that, apply the migration and restart the connector.
Then finish the installation.
The installation should succeed and the provided API key should be stored in the database.
For the complete code see tag [**step5.1**](https://github.com/connctd/giphy-connector/tree/step5.1).

Now that we have one API key per installation, we should start using it.
Instead of setting the API key once at start up, we will change the Giphy provider to retrieve the key from the installation and retrieve the new random gif with this key.

**giphy.go**

```golang
func (h *Handler) Run() {
	for {
		h.addNewInstances()

		for _, instance := range h.instances {
			if err := h.setApiKey(instance.InstallationID); err != nil {
				logrus.WithError(err).Errorln("failed to set API key for " + instance.InstallationID)
				continue
			}

			randomGif, err := h.getRandomGif()
			if err != nil {
				continue
			}

			h.updateChannel <- connector.GiphyUpdate{
				InstanceId: instance.ID,
				Value:      randomGif,
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func (h *Handler) setApiKey(installationId string) error {
	installation, ok := h.installations[installationId]
	if !ok {
		return errors.New("installation not registered")
	}
	key, ok := installation.GetConfig("giphy_api_key")
	if !ok {
		return errors.New("could not find api key")
	}

	h.giphyClient.APIKey = key.Value
	return nil
}
```

As we see, the Giphy provider needs access to the installations to be able to retrieve the key from the configuration.
This means that we have to register the installations with the provider, like we did before with the instances.
Also the provider needs to now, which instance belongs to which installation.
We solve the seconds problem by registering instances instead of only registering their IDs, which will give us the instances installation ID.

**giphyhandler.go**

```golang
type Provider interface {
	UpdateChannel() <-chan GiphyUpdate
	RegisterInstances(instances ...*Instance) error
	RegisterInstallations(installations ...*Installation) error
}
```

We will register the installations at start up and whenever a new installation is created.

**protocol.go**

```golang
func (g *GiphyConnector) init() {
	installations, err := g.db.GetInstallations(context.Background())
	if err != nil {
		g.logger.WithError(err).Error("Failed to retrieve instances from db")
		return
	}
	g.giphyProvider.RegisterInstallations(installations...)

    [...]

	g.giphyProvider.RegisterInstances(instances...)

	go g.giphyEventHandler(context.Background())
}

// AddInstallation is called by the HTTP handler when it retrieved an installation request
func (g *GiphyConnector) AddInstallation(ctx context.Context, installationRequest connector.InstallationRequest) (*connector.InstallationResponse, error) {
    [...]

	g.giphyProvider.RegisterInstallations(&giphy.Installation{
		ID:            installationRequest.ID,
		Token:         installationRequest.Token,
		Configuration: installationRequest.Configuration,
	})

	return nil, nil
}
```

As you can see, this required us to implement a Go representation for installations and a database method to retrieve the installations.
Take a look at **model.go**, **mysql.go** and **service.go** for more details.
We have also implemented a convenience method to get configuration parameters by ID from an installation, which allows us to easily retrieve the API key from an installation.

As a last step, we can now also remove the API key from **runs.sh** and **main.go**.

For the complete code see tag [**step5.2**](https://github.com/connctd/giphy-connector/tree/step5.2).

## Action Requests
