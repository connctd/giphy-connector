package main

import (
	"github.com/connctd/connector-go"
	"github.com/connctd/connector-go/models"
)

const (
	RandomComponentId       = "random"
	RandomPropertyId        = "value"
	SearchComponentId       = "search"
	SearchPropertyId        = "value"
	SearchActionId          = "search"
	SearchActionParameterId = "keyword"
)

// buildThing returns a thing that can be registered with the connctd platform.
// Note that the thing ID is generated by connctd and returned when the thing is created.
// The connctd platform will store all information regarding the thing.
// The connector therefore should only store its ID.
// The thing will have to components.
// randomComponent will periodically updated by a new random value.
// searchComponent will only be updated when a search action is triggered.
func thingTemplate(request connector.InstantiationRequest) []models.Thing {
	return []models.Thing{
		{
			Name:            "Giphy",
			Manufacturer:    "IoT connctd GmbH",
			DisplayType:     "core.SENSOR",
			MainComponentID: RandomComponentId,
			Status:          "AVAILABLE",
			Attributes:      []models.ThingAttribute{},
			Components: []models.Component{
				{
					ID:            RandomComponentId,
					Name:          "Giphy random component",
					ComponentType: "core.Sensor",
					Capabilities: []string{
						"core.MEASURE",
					},
					Properties: []models.Property{
						{
							ID:    RandomPropertyId,
							Name:  "Giphy random property",
							Value: "",
							Type:  models.ValueTypeString,
						},
					},
					Actions: []models.Action{},
				},
				{
					ID:            SearchComponentId,
					Name:          "Giphy search",
					ComponentType: "core.Sensor",
					Capabilities: []string{
						"core.SEARCH",
					},
					Properties: []models.Property{
						{
							ID:   SearchPropertyId,
							Name: "Giphy search property",
							Type: models.ValueTypeString,
						},
					},
					Actions: []models.Action{
						{
							ID:   SearchActionId,
							Name: "Giphy search action",
							Parameters: []models.ActionParameter{
								{
									Name: SearchActionParameterId,
									Type: models.ValueTypeString,
								},
							},
						},
					},
				},
			},
		},
	}
}
