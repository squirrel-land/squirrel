package models

import (
	"github.com/songgao/squirrel/models/common"
	"math"
	"math/rand"
)

type september1st struct {
	positionManager    *common.PositionManager
	noDeliveryDistance float64
}

func newSeptember1st() common.September {
	return &september1st{}
}

func (september *september1st) ParametersHelp() string {
	return `September1st delivers packets only based on distance between nodes. It applies
a packet loss (d/D)^4 to each packet, where d is the distance between the two
nodes, and D is the maximum communication range. It does not consider
interference.

  "LowestZeroPacketDeliveryDistance": float64, required;
                                      Maximum transmission range.
    `
}

func (september *september1st) Configure(config map[string]interface{}) (err error) {
	dist, ok := config["LowestZeroPacketDeliveryDistance"].(float64)
	if ok != true {
		return ParametersNotValid
	}
	september.noDeliveryDistance = dist
	return nil
}

func (september *september1st) Initialize(positionManager *common.PositionManager) {
	september.positionManager = positionManager
}

func (september *september1st) SendUnicast(source int, destination int, size int) bool {
	return september.isToBeDelivered(source, destination)
}

func (september *september1st) SendBroadcast(source int, size int, underlying []int) []int {
	count := 0
	for _, i := range september.positionManager.Enabled() {
		if i != source && september.isToBeDelivered(source, i) {
			underlying[count] = i
			count++
		}
	}
	return underlying[:count]
}

func (september *september1st) isToBeDelivered(id1 int, id2 int) bool {
	if september.positionManager.IsEnabled(id1) && september.positionManager.IsEnabled(id2) {
		dist := september.positionManager.Distance(id1, id2)
		if dist < september.noDeliveryDistance*0.8 {
			return true
		}
		return rand.Float64() > math.Pow(dist/september.noDeliveryDistance, 4)
	} else {
		return false
	}
}
