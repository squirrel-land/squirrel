package simpleModels

import (
	"../master"
	"math"
	"math/rand"
)

type september1st struct {
	nodes              []*master.Position
	noDeliveryDistance float64
}

func NewSeptember1st() master.September {
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

func (september *september1st) Initialize(nodes []*master.Position) {
	september.nodes = nodes
}

func (september *september1st) SendUnicast(source int, destination int) bool {
	return september.isToBeDelivered(source, destination)
}

func (september *september1st) SendBroadcast(source int, underlying []int) []int {
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
	p1 := september.nodes[id1]
	p2 := september.nodes[id2]
	if p1 == nil || p2 == nil {
		return false
	}
	dist := distance(p1, p2)
	r := rand.Float64()
	return r > math.Pow(dist/september.noDeliveryDistance, 4)
}

func distance(p1 *master.Position, p2 *master.Position) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2) + math.Pow(p1.Height-p2.Height, 2))
}
