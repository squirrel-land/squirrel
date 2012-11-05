package simpleModels

import (
	"../master"
)

type SimpleMobilityManager struct {
	nodes []*master.Position
}

func NewSimpleMobilityManager(config map[string]interface{}) (master.MobilityManager, error) {
	return &SimpleMobilityManager{}, nil
}

func (mobilityManager *SimpleMobilityManager) GenerateNewNode() *master.Position {
	return &master.Position{0, 0, 0}
}

func (mobilityManager *SimpleMobilityManager) SetMobileNodesSlice(nodes []*master.Position) {
	mobilityManager.nodes = nodes
}
