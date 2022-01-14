# Connector Tutorial

In this tutorial we are going to implement a connector to connect [Giphy](https://giphy.com/) to our platform.
Source code for this tutorial can be found at [Github](https://github.com/connctd/giphy-connector).

## Requirements

The connector connects Giphy to the connctd platform.
Therefore you need an account for the [connctd Developer Center](https://devcenter.connctd.io/).

We are using the Giphy API which requires an account with Giphy and a Giphy API key.
See the [Giphy documentation](https://developers.giphy.com/docs/api#quick-start-guide) on how to acquire them.

We will implement the connector in [Go](https://golang.org/) and assume basic knowledge of the language.
Using Go allows us to use the [connector-go](https://github.com/connctd/connector-go) SDK provided by connctd which will simplify the implementation and take care of most boilerplate code.

The connector must be reachable from the connctd platform.
For this tutorial we use [ngrok](https://ngrok.com/) to create public URLs for our local connector.
While this is not a production ready solution, it gives us a simple development environment and lets us quickly deploy changes to our connector.
In a production environment, you can use a hosting provider of your choice.

Note that the code examples provided in the tutorial leave out a lot of error handling for clarity.
Check out the [code repository](https://github.com/connctd/giphy-connector) for the complete code.

## Connector Concept

Conceptually a connector consists of two parts:

1) A HTTP handler, that provides endpoints for the connector protocol

    The connector protocol defines a set of callbacks that are called whenever an installation or instance is created or removed and when an action is triggered on the connctd platform.
    The callbacks URL can be defined when the connector is published at the connctd platform and new actions can be registered as part of a thing.

2) A service that implements the technology we want to connect to the connctd platform and communicates to the connctd platform via API provided by the platform.

    Connectors use the thing concept to represent a technology at the connctd platform.
    Things can contain a variety of developer defined components, properties and actions and while the developer decides which things get created, all things will reside at the connctd platform.
    The connector protocol defines an API that can be used to manage and update the created things and actions.

For more information on the connector concept and an in-depth description of the connector protocol, take a look at our [documentation](https://docs.connctd.io/connector/connectors/).
For a better understanding on how connectors work we recommend to read at least the introduction to connectors and the connector protocol but it is not strictly necessary to follow this tutorial.

## Connector SDK

The endpoints implemented by the handler are defined by the protocol and are fully implemented by the SDK.
Developers should therefore never need to implement the handler on their own.
The second part is further divided by the SDK in a *service* and *provider*.
The *service* implements an interface that is consumed by the handler and processes all requests from the connctd platform.
The SDK implements a default service that manages the bookkeeping needed for most connectors.
It will save new installations and instances, create new things at the connctd platform, forward action requests to the provider and listen to update events from the provider.
Using the default service should be sufficient for most connectors and it can easily adapted to more complex use cases.
We will use the default service with the default database implementation, which means that we can focus on the provider.

## Provider

The SDK defines an interface that must be implemented by all providers to support the service definition.


```golang
type Provider interface {
	UpdateChannel() <-chan UpdateEvent
	RequestAction(ctx context.Context, instance *Instance, actionRequest ActionRequest) (restapi.ActionRequestStatus, error)

	RegisterInstallations(installations ...*Installation) error
	RemoveInstallation(installationId string) error

	RegisterInstances(instances ...*Instance) error
	RemoveInstance(instanceId string) error

}

```

The interface allows the service to register and remove new installations and instances, but more importantly it provides an update channel and a method to request actions from the provider.
The default service will listen on the update channel and propagate received updates to the connctd platform.
The provider therefore does not need to implement anything specific to connctd.
The action request method allows the provider to react to action requests.
It can either execute the action directly or asynchronously by returning a pending status.
The registration methods give the provider access to installations and instances.
This allows the provider to access configuration parameters specific to a installation or instance.

The SDK contains a default provider which handles registrations and provides an update channel and an asynchronous action handler.
The default provider is meant to be embedded and extended by the connector and all methods can easily be overwritten if they do not fit a specific connector.

Using the default implementations, developing a new connector breaks down to two basic tasks:

1. Defining the things we want to create to represent the technology
2. Implementing the features specific to the technology

This tutorial implements two features of the Giphy platform.
The first feature is a periodic update of a thing with a new random gif from Giphy.
Second we want to enable a simple keyword search via the Giphy API.

### Thing Abstraction

We can represent this features with a single thing with two components and one property each.
The first component will contain the random gif and is updated periodically by the provider.
The second component will contain the search result and is only updated when a search action is triggered with a keyword.
Therefore we also define the action in the second component.

In general things should be created for every new instance, because instances represent the end user of a connector.
The default service can create things for us whenever a new instance is created when we provide a `ThingTemplates` callback to it.
The function is called whenever a new instance is created and the service will register all things returned by the function. Note that things of course can also be created at a later point of time of the instance lifecylce.

```golang
type ThingTemplates func(request InstantiationRequest) []restapi.Thing

func thingTemplate(request connector.InstantiationRequest) []restapi.Thing {
	return []restapi.Thing{
		{
			Name:            "Giphy",
			Manufacturer:    "IoT connctd GmbH",
			DisplayType:     "core.SENSOR",
			MainComponentID: RandomComponentId,
			Status:          "AVAILABLE",
			Attributes:      []restapi.ThingAttribute{},
			Components: []restapi.Component{
				{
					ID:            RandomComponentId,
					Name:          "Giphy random component",
					ComponentType: "core.Sensor",
					Capabilities: []string{
						"core.MEASURE",
					},
					Properties: []restapi.Property{
						{
							ID:    RandomPropertyId,
							Name:  "Giphy random property",
							Value: "",
							Type:  restapi.ValueTypeString,
						},
					},
					Actions: []restapi.Action{},
				},
				{
					ID:            SearchComponentId,
					Name:          "Giphy search",
					ComponentType: "core.Sensor",
					Capabilities: []string{
						"core.SEARCH",
					},
					Properties: []restapi.Property{
						{
							ID:   SearchPropertyId,
							Name: "Giphy search property",
							Type: restapi.ValueTypeString,
						},
					},
					Actions: []restapi.Action{
						{
							ID:   SearchActionId,
							Name: "Giphy search action",
							Parameters: []restapi.ActionParameter{
								{
									Name: SearchActionParameterId,
									Type: restapi.ValueTypeString,
								},
							},
						},
					},
				},
			},
		},
	}
}

```

The function returns a thing with the component as we have discussed above.
Note that the search action also contains a required parameter, which we will use as the keyword.
We also store some of the component and property ID in constants so that we can easily reference them later on.
In our case, we don't use anything from the provided instantiation requests, but other connectors might use per instance configurations.

### Giphy Connector

Now lets implement the provider with both features.
We start by defining a new struct containing the default provider.
Note that the new struct implements the provider interface by embedding the default provider.
To communicate with the Giphy API we will also import and use an [API client](https://github.com/peterhellberg/giphy).
For the periodic update we will periodically iterate over all instances, call the Giphy API to receive a new random gif and publish a update event on our event channel.
The default provider implements two additional methods which we use here.
The `Update()` method will add all newly registered installations and instances to our provider (and remove old ones) and the `UpdateEvent()` method will push an event to the underlying event channel.
Since the default service is listening to the event channel and handle the property update, this is all we have to implement for this feature.
Note that the `ThingMapping` is automatically created for us by the default service when new instances are created.
Also we set a API key of the Giphy client for every installations.
We will explain later, how we can use a different key per installation using configuration parameters.

```golang
type GiphyProvider struct {
	provider.DefaultProvider
	giphyClient *giphyClient.Client
}

// periodicUpdate starts an endless loop which will periodically update the random component of each instance
func (h *GiphyProvider) periodicUpdate() {
	for {
		h.Update()

		for _, instance := range h.Instances {
			randomGif, err := h.getRandomGif(instance)
			update := connector.UpdateEvent{
				PropertyUpdateEvent: &connector.PropertyUpdateEvent{
					InstanceId:  instance.ID,
					ThingId:     instance.ThingMapping[0].ThingID,
					ComponentId: RandomComponentId,
					PropertyId:  RandomPropertyId,
					Value:       randomGif,
				},
			}
			h.UpdateEvent(update)
		}
		time.Sleep(TIME_BETWEEN_UPDATES)
	}
}


// getRandomGif uses the Giphy API to return a new random gif.
func (h *GiphyProvider) getRandomGif(instance *connector.Instance) (string, error) {
	h.setApiKey(instance.InstallationID)
	random, err := h.giphyClient.Random([]string{})
	return random.Data.URL, nil
}

func (h *GiphyProvider) setApiKey(installationId string) error {
	installation, ok := h.Installations[installationId]
	key, ok := installation.GetConfig("giphy_api_key")

	h.giphyClient.APIKey = key.Value
	return nil
}

```

Next we can implement the search action.
The default provider already implements a action handler which will push all actions to the action channel and return a pending status to the service.
With this in place, the connector can listen to the action channel, execute the search and push an property update and action event to the event channel.
If the service receives an event containing both event types it will handle the property update first and then update the action state if the update was successful.
Otherwise it will update the action state to failed.

```golang
func (h *GiphyProvider) actionHandler() {
	for pendingAction := range h.ActionChannel() {
		update := connector.UpdateEvent{
			ActionEvent: &connector.ActionEvent{
				InstanceId: pendingAction.Instance.ID,
				RequestId:  pendingAction.ID,
				Response:   &connector.ActionResponse{},
			},
		}

		switch pendingAction.ActionID {
		case "search":
			keyword := pendingAction.Parameters["keyword"]
			result, err := h.getSearchResult(pendingAction.Instance, keyword)

			if err != nil {
				update.ActionEvent.Response = &connector.ActionResponse{
					Status: restapi.ActionRequestStatusFailed,
					Error:  err.Error(),
				}
				h.UpdateEvent(update)
				continue
			}

			update.ActionEvent.Response = &connector.ActionResponse{
				Status: restapi.ActionRequestStatusCompleted,
			}
			update.PropertyUpdateEvent = &connector.PropertyUpdateEvent{
				ThingId:     pendingAction.Instance.ThingMapping[0].ThingID,
				InstanceId:  pendingAction.Instance.ID,
				ComponentId: SearchComponentId,
				PropertyId:  SearchPropertyId,
				Value:       result,
			}
			h.UpdateEvent(update)
	}
}

// getSearchResult uses the Giphy API to search for the given keyword.
func (h *GiphyProvider) getSearchResult(instance *connector.Instance, keyword string) (string, error) {
	h.setApiKey(instance.InstallationID)

	h.giphyClient.Limit = 1
	result, err := h.giphyClient.Search([]string{keyword})
	if err != nil {
		return "", err
	}
	if len(result.Data) <= 0 {
		return "", errors.New("no search result found")
	}

	return result.Data[0].URL, nil
}

```

Since both methods should run for the whole runtime of the connector, we also implement a little helper function that just runs both methods in a goroutine.

### Connector Publication

Now it's time to stich everything together and run the connector.
In our main method, we will instantiate the provider, the default service and database and the connector handler for the HTTP endpoints.
Last we run the event handler, the action handler and periodic update and serve the HTTP handler.
Note that the HTTP handler also requires a public key to authenticate all requests from the connctd platform.
We will receive the key when we publish the connector at the platform.
Also, we need a database to save all new installations and instances.
This example will use a Sqlite database which doesn't need any setup.
See the full code on how to migrate the database and how to use a different database.

```golang
func main() {
	// Create the Giphy provider
	giphyProvider := NewGiphyProvider()

	// Create a new instance of our connector
	dbClient, err := db.NewDBClient(db.DefaultOptions, connector.DefaultLogger)

	// Create a new client for the connctd API
	connctdClient, err := connector.NewClient(nil, connector.DefaultLogger)

	// Create a new instance of our connector
	service, err := service.NewConnectorService(dbClient, connctdClient, giphyProvider, thingTemplate, connector.DefaultLogger)

	// Create a new HTTP handler using the service
	httpHandler := connector.NewConnectorHandler(nil, service, publicKey)

	// Start the event handler listening to action and property update events
	service.EventHandler(ctx)

	// Start Giphy provider
	giphyProvider.Run()

	// Start the http server using our handler
	http.ListenAndServe(":8080", httpHandler)
}
```

You can now run the connector with the provided `run.sh` script.
On the first run, you should add `-migrate` to the run command in the script.
To create a public URL which can be reached from the connctd platform, use ngrok with `ngrok http 8080`.
It should print an URL prefixed with https which we can use to publish our connector.

To be able to communicate with the connctd platform we first have to publish the connector with the connctd platform.
To publish the connector, login to the [Developer Center](https://devcenter.connctd.io/) and create a new App, if you haven't already.
Make sure to safely store the client id and client secret, we will need it later on and it can not be restored.
Next go to the [Connector Store](https://devcenter.connctd.io/connectors) and click "Create Connector".

In the second step, provide the callback URLs:

installationCallbackUrl: https://-ngrokId-.ngrok.io/installations

instanceCallbackUrl: https://-ngrokId-.ngrok.io/instantiations

actionCallbackUrl: https://-ngrokId-.ngrok.io/actions

In step 3, add a new entry to the installation configuration with the ID ***giphy_api_key***, a value type of ***STRING(Text)*** and required set to true.
This instructs the connctd platform to ask for a configuration parameter for every installation and the parameters will be send to the connector with the installation request.
The default service will automatically save all configuration parameters and provide them as part of the registration.

Complete the publication and make sure to save the public key which is shown after the publication.
Before we install the connector, we have to provide the public key to the connector.
To do this take a look at **run.sh** and replace the key exported for GIPHY_CONNECTOR_PUBLIC_KEY with the new key.
After this, restart the connector (do not restart ngrok).

You can now install the connector.
Select your app in the Developer Center sidebar then go to the Connector Store and click "Install from Connector Store".
You should see the connector listed under your private connectors.
Click on it and click the install button.
The installation should ask you for the API key and succeed if you provide one. You have now successfully installed the giphy connector for your app which means on your connector side a installation should have been added. You can now create connector instances for every of your endcustomers using your app.

Creating an instance is not possible from within the Developer Center, since it is usually done programmatically by the App using a connector.
We can however use the connctd GraphQL API to create an instance manually.
For this you need the installation ID, which you can find in the [Developer Center](https://devcenter.connctd.io/connectors).
Click on your installation and the ID is shown.
Also you need an authorization token which you can acquire with the [Client Credential Flow](https://docs.connctd.io/general/oauth2/#client-credentials-flow).
Alternatively you can use the [GraphQL Explorer](https://docs.connctd.io/graphql/explorer) with your client credentials to instantiate the connector.

To instantiate the connector send the following request, replacing your auth token and installation ID.

```http
curl --location --request POST 'https://api.connctd.io/api/v1/query' \
--header 'Authorization: Bearer --yourAuthToke--' \
--header 'Content-Type: application/json' \
--data-raw '{"query":"mutation InstantiateConnector {\n    instantiateConnector(installationId: \"--installtionID--\", configuration: []\n    ) {\n        id\n        installation {\n            id\n        }\n        state\n        stateDetails\n    }\n}","variables":{}}'
```

As soon as the instance was created, the connector should start updating the random component.
You should be able to see the created things and the property updates.
You can also trigger the action with the [GraphQL API](https://docs.connctd.io/graphql/things/#trigger-actions).