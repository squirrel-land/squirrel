package models

import (
	"github.com/songgao/squirrel/models/common"
)

type september0th struct {
	positionManager *common.PositionManager
}

func newSeptember0th() common.September {
	return &september0th{}
}

func (september *september0th) ParametersHelp() string {
	return `September0th delivers every packet sent into squirrel as long as the src and dst are valid.`
}

func (september *september0th) Configure(config map[string]interface{}) (err error) {
	return nil
}

func (september *september0th) Initialize(positionManager *common.PositionManager) {
	september.positionManager = positionManager
}

func (september *september0th) SendUnicast(source int, destination int, size int) bool {
	return september.isToBeDelivered(source, destination)
}

func (september *september0th) SendBroadcast(source int, size int, underlying []int) []int {
	count := 0
	for _, i := range september.positionManager.Enabled() {
		if i != source {
			underlying[count] = i
			count++
		}
	}
	return underlying[:count]
}

func (september *september0th) isToBeDelivered(id1 int, id2 int) bool {
	if september.positionManager.IsEnabled(id1) && september.positionManager.IsEnabled(id2) {
		return true
	} else {
		return false
	}
}
