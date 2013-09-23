package models

import (
	"github.com/songgao/squirrel/models/common"
)

var MobilityManagers = map[string]func() common.MobilityManager{
	"StaticUniformPositions": newStaticUniformPositions,
	"StaticDefinedPositions": newStaticDefinedPositions,
	"InteractivePositions":   newInteractivePositions,
}

var Septembers = map[string]func() common.September{
	"September0th": newSeptember0th,
	"September1st": newSeptember1st,
	"September2nd": newSeptember2nd,
}
