package models

import (
	"github.com/songgao/squirrel/models/common"
	"math"
	"math/rand"
)

type september1st struct {
	nodes              []*common.Position
	noDeliveryDistance float64
}

func NewSeptember1st() common.September {
	return &september1st{}
}

func (september *september1st) ParametersHelp() string {
	return ""
}

func (september *september1st) Configure(config map[string]interface{}) (err error) {
	dist, ok := config["LowestZeroPacketDeliveryDistance"].(float64)
	if ok != true {
		return ParametersNotValid
	}
	september.noDeliveryDistance = dist
	return nil
}

func (september *september1st) Initialize(nodes []*common.Position) {
	september.nodes = nodes
}

func (september *september1st) SendUnicast(source int, destination int, size int) bool {
	return september.isToBeDelivered(source, destination)
}

func (september *september1st) SendBroadcast(source int, size int, underlying []int) []int {
	count := 0
	for i := 1; i < len(september.nodes); i++ {
		if i != source && september.isToBeDelivered(source, i) {
			underlying[count] = i
			count++
		}
	}
	return underlying[:count]
}

func (september *september1st) isToBeDelivered(id1 int, id2 int) bool {
	september.nodes[id1].Mu.RLock()
	september.nodes[id2].Mu.RLock()
	defer september.nodes[id1].Mu.RUnlock()
	defer september.nodes[id2].Mu.RUnlock()

	p1 := september.nodes[id1]
	p2 := september.nodes[id2]
	if p1 == nil || p2 == nil {
		return false
	}
	dist := distance(p1, p2)
	if dist < september.noDeliveryDistance*0.8 {
		return true
	}
	return rand.Float64() > math.Pow(dist/september.noDeliveryDistance, 4)
}
