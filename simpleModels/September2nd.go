package simpleModels

import (
	"../modelDep"

	"math"
	"math/rand"
	"sort"
    "time"
)

type september2nd struct {
	noDeliveryDistance float64

	nodes      []*modelDep.Position
	buckets    []leakyBucket // measured by bytes of data; if non-maxDataRate is used, value added into buckets is the number of bytes that could be sent on maxDataRate within the same amount of time
	dataRates   []float64 // supported data rates in increasing order
    maxDataRate float64
}

func NewSeptember2nd() modelDep.September {
	ret := new(september2nd)
	ret.dataRates = []float64{6, 9, 12, 18, 24, 36, 48, 54}
    ret.maxDataRate = ret.dataRates[len(ret.dataRates)-1] * 1024 * 1024
    return ret
}

func (september *september2nd) ParametersHelp() string {
	return ""
}

func (september *september2nd) Configure(config map[string]interface{}) (err error) {
	dist, okDist := config["LowestZeroPacketDeliveryDistance"].(float64)

	if true != (okDist) {
		return ParametersNotValid
	}

	september.noDeliveryDistance = dist

	return nil
}

func (september *september2nd) Initialize(nodes []*modelDep.Position) {
	september.nodes = nodes
	september.buckets = make([]leakyBucket, len(nodes), len(nodes))
	for it := range september.buckets {
        september.buckets[it] = &leakyBucket{
            BucketSize: int32(1024 * 1024 * september.maxDataRate),
            OutResolution: 10 * time.Millisecond,
            OutPerSecond: int32(1024 * 1024 * september.maxDataRate)
        }
        september.buckets[it].Go()
	}
}

func (september *september2nd) SendUnicast(source int, destination int) bool {
	return true
}

func (september *september2nd) SendBroadcast(source int, underlying []int) []int {
	return nil
}

func (september *september2nd) tryToDeliver(id1 int, id2 int) bool {
	return true
}

func (september *september2nd) dataRate(src, dst int) float64 {
    return september.dataRate[len(september.datarate) - 1]
}
