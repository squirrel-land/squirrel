package master

type September1st struct {
	nodes []*Position
}

func (september *September1st) SetMobileNodesSlice(nodes []*Position) {
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
