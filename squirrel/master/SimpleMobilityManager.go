package master

type SimpleMobilityManager struct {
	nodes []*Position
}

func (mobilityManager *SimpleMobilityManager) GenerateNewNode() *Position {
	return &Position{0, 0, 0}
}

func (mobilityManager *SimpleMobilityManager) SetMobileNodesSlice(nodes []*Position) {
	mobilityManager.nodes = nodes
}
