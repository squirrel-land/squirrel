package models

import (
	"github.com/songgao/squirrel/models/common"
)

var MobilityManagers = map[string]func() common.MobilityManager{
	"StaticUniformPositions": newStaticUniformPositions,
	"StaticDefinedPositions": newStaticDefinedPositions,
}

var Septembers = map[string]func() common.September{
	"September1st": newSeptember1st,
	"September2nd": newSeptember2nd,
}
