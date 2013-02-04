package main

import (
	"github.com/songgao/squirrelModels"
)

var mobilityManagers = map[string]typeMobilityManagerConstructor{
	"StaticUniformPositions": simpleModels.NewStaticUniformPositions,
}

var septembers = map[string]typeSeptemberConstructor{
	"September1st": simpleModels.NewSeptember1st,
	"September2nd": simpleModels.NewSeptember2nd,
}
