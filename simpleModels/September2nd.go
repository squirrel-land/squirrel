package simpleModels

import (
	"../modelDep"

	"math"
	"math/rand"
)

type september2nd struct {
	noDeliveryDistance float64
	interferenceRange  float64

	nodes       []*modelDep.Position
	buckets     []*leakyBucket // measured by bytes of data; if non-maxDataRate is used, value added into buckets is the number of bytes that could be sent on maxDataRate within the same amount of time
	dataRates   []float64      // supported data rates in Mbps in increasing order
	maxDataRate float64
}

func NewSeptember2nd() modelDep.September {
	ret := new(september2nd)
	ret.dataRates = []float64{6, 9, 12, 18, 24, 36, 48, 54}
	ret.maxDataRate = ret.dataRates[len(ret.dataRates)-1]
	return ret
}

func (september *september2nd) ParametersHelp() string {
	return ""
}

func (september *september2nd) Configure(config map[string]interface{}) (err error) {
	dist, okDist := config["LowestZeroPacketDeliveryDistance"].(float64)
	iRange, okIRange := config["InterferenceRange"].(float64)

	if true != (okDist && okIRange) {
		return ParametersNotValid
	}

	september.noDeliveryDistance = dist
	september.interferenceRange = iRange

	return nil
}

func (september *september2nd) Initialize(nodes []*modelDep.Position) {
	september.nodes = nodes
	september.buckets = make([]*leakyBucket, len(nodes), len(nodes))
	for it := range september.buckets {
		september.buckets[it] = &leakyBucket{
			BucketSize:        int32(1024 * 1024 * september.maxDataRate),
			OutResolution:     int32(10),
			OutPerMilliSecond: int32(1024 * 1024 * september.maxDataRate / 1000),
		}
		september.buckets[it].Go()
	}
}

func (september *september2nd) SendUnicast(source int, destination int, size int) bool {
	p1 := september.nodes[source]
	p2 := september.nodes[destination]
	if p1 == nil || p2 == nil {
		return false
	}
	dist := distance(p1, p2)

	// Go through source bucket
	if !september.buckets[source].In(size) {
		return false
	}

	// Since the packet is out in the air, interference should be put on neighbor nodes
	for i := 1; i < len(september.nodes); i++ {
		if i != source && i != destination && dist < september.interferenceRange {
			september.buckets[i].In(size)
		}
	}

	// The packet takes the adventure in the air (fading, etc.)
	if dist > september.noDeliveryDistance*0.85 {
		if rand.Float64() < math.Pow(dist/september.noDeliveryDistance, 5) {
			return false
		}
	}

	// Go through destination bucket
	if !september.buckets[destination].In(size) {
		return false
	}

	// The packet is gonna be delivered!
	return true
}

func (september *september2nd) SendBroadcast(source int, size int, underlying []int) []int {
	p1 := september.nodes[source]
	if p1 == nil {
		return underlying[:0]
	}

	// Go through source bucket
	if !september.buckets[source].In(size) {
		return underlying[:0]
	}

	count := 0
	for i := 1; i < len(september.nodes); i++ {
		p2 := september.nodes[i]
		if p2 == nil {
			continue
		}
		dist := distance(p1, p2)

		// Since the packet is out in the air, interference should be put on neighbor nodes
		if dist < september.interferenceRange {
			if !september.buckets[i].In(size) {
				// Go through destination bucket. If rejected by the bucket, the broadcasted packet should not be delivered to this node
				continue
			}
		}

		// The packet takes the adventure in the air (fading, etc.). There's still a possibility the packet is not delivered to this node
		if dist > september.noDeliveryDistance*0.85 {
			if rand.Float64() < math.Pow(dist/september.noDeliveryDistance, 5) {
				continue
			}
		}

		// The packet is gonna be delivered!
		underlying[count] = i
		count++
	}
	return underlying[:count]
}

func (september *september2nd) dataRate(src, dst int) float64 {
	return september.dataRates[len(september.dataRates)-1]
}
