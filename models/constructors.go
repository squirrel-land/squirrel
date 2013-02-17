package models

import (
	"github.com/songgao/squirrel/models/common"
)

var MobilityManagers = map[string]func() common.MobilityManager{
	"StaticUniformPositions": NewStaticUniformPositions,
}

var Septembers = map[string]func() common.September{
	"September1st": NewSeptember1st,
	"September2nd": NewSeptember2nd,
}
