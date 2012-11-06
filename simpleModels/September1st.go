package simpleModels

import (
	"../master"
)

type September1st struct {
	nodes []*master.Position
}

func NewSeptember1st() master.September {
	return &September1st{}
}

func (september *September1st) ParametersHelp() string {
	return ""
}

func (september *September1st) Configure(config map[string]interface{}) (err error) {
	return nil
}

func (september *September1st) Initialize(nodes []*master.Position) {
	september.nodes = nodes
}

func (september *September1st) SendUnicast(source int, destination int) bool {
	return true
}

func (september *September1st) SendBroadcast(source int, underlying []int) []int {
	count := 0
	for i := 1; i < len(september.nodes); i++ {
		if september.nodes[i] != nil {
			underlying[count] = i
			count++
		}
	}
	return underlying[:count]
}
