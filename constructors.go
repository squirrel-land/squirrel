package main

import (
	"./simpleModels"
)

var mobilityManagers = map[string]typeMobilityManagerConstructor{
	"SimpleMobilityManager": simpleModels.NewSimpleMobilityManager,
}

var septembers = map[string]typeSeptemberConstructor{
	"September1st": simpleModels.NewSeptember1st,
}
