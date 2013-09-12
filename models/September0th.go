package models

import (
	"github.com/songgao/squirrel/models/common"
)

type september0th struct {
	nodes              []*common.Position
	noDeliveryDistance float64
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

func (september *september0th) Initialize(nodes []*common.Position) {
	september.nodes = nodes
}

func (september *september0th) SendUnicast(source int, destination int, size int) bool {
	return september.isToBeDelivered(source, destination)
}

func (september *september0th) SendBroadcast(source int, size int, underlying []int) []int {
	count := 0
	for i := 1; i < len(september.nodes); i++ {
		if i != source && september.isToBeDelivered(source, i) {
			underlying[count] = i
			count++
		}
	}
	return underlying[:count]
}

func (september *september0th) isToBeDelivered(id1 int, id2 int) bool {
	if september.nodes[id1] == nil || september.nodes[id2] == nil {
		return false
	} else {
		return true
	}
}
