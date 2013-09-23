package models

import (
	"github.com/songgao/squirrel/models/common"

	"math"
	"math/rand"
)

type september2nd struct {
	noDeliveryDistance float64
	interferenceRange  float64

	positionManager *common.PositionManager
	buckets         []*leakyBucket // measured by number of nanoseconds used;
	dataRate        float64        // bit data rates in bit/nanosecond

	// MAC layer time in nanoseconds
	difs    int
	sifs    int
	slot    int
	cWindow int // assuming fixed MAC layer contention window

	// MAC layer frame properties in bits
	macFrameMaxBody  int
	macFrameOverhead int
}

func newSeptember2nd() common.September {
	ret := new(september2nd)
	ret.dataRate = 54 * 1024 * 1024 * 1e-9 // 54 Mbps
	ret.slot = 9e3                         // 9 microseconds
	ret.sifs = 10e3                        // 10 microseconds
	ret.difs = ret.sifs + 2*ret.slot       // 28 microseconds
	ret.cWindow = 127                      // 127 slots
	ret.macFrameMaxBody = 2312 * 8
	ret.macFrameOverhead = 34
	return ret
}

func (september *september2nd) ParametersHelp() string {
	return `
September2nd delivers packets based on a near 802.11 Ad-hoc model. It considers
distance between nodes and interference, etc..

  "LowestZeroPacketDeliveryDistance": float64, required;
                                      Maximum transmission range.
  "InterferenceRange":                float64, required;
                                      Maximum interference range, normally
                                      slighly larger than 2x transmission range.
    `
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

func (september *september2nd) Initialize(positionManager *common.PositionManager) {
	september.positionManager = positionManager
	september.buckets = make([]*leakyBucket, positionManager.Capacity())
	for it := range september.buckets {
		september.buckets[it] = &leakyBucket{
			BucketSize:        int32(1000 * 1000 * 1000), // 1 second
			OutResolution:     int32(10),                 //10 millisecond
			OutPerMilliSecond: int32(1000 * 1000),        // 1000 microseconds gone every milliscond
		}
		september.buckets[it].Go()
	}
}

func (september *september2nd) nanosecByData(bytes int) int {
	framebody := int(float64(bytes*8) / september.dataRate)
	frameOverhead := september.macFrameOverhead * (bytes*8/september.macFrameMaxBody + 1)
	return framebody + frameOverhead
}

func (september *september2nd) cw() int {
	return september.slot * september.cWindow / 2
}

func (september *september2nd) nanosecByPacket(packetSize int) int {
	return september.difs + september.cw() + september.nanosecByData(packetSize) + september.sifs + september.nanosecByData(0) // the last nanosecByData(0) is for MAC layer ACK
}

func (september *september2nd) ackIntererence() int {
	return september.difs + september.cw() + september.sifs + september.nanosecByData(0)
}

func (september *september2nd) deliverRate(dest int, dist float64) float64 {
	usage := september.buckets[dest].Usage()
	p_rate := (1-usage)*.1 + .9 // usage transformed from [0, 1] to [.9, 1]
	return p_rate * (1 - math.Pow(dist/september.noDeliveryDistance, 3))
}

func (september *september2nd) SendUnicast(source int, destination int, size int) bool {
	if !(september.positionManager.IsEnabled(source) && september.positionManager.IsEnabled(destination)) {
		return false
	}

	// Go through source bucket
	if !september.buckets[source].In(september.nanosecByPacket(size)) {
		return false
	}

	dist := september.positionManager.Distance(source, destination)

	// Since the packet is out in the air, interference should be put on neighbor nodes
	for _, i := range september.positionManager.Enabled() {
		d1 := september.positionManager.Distance(source, i)
		d2 := september.positionManager.Distance(destination, i)
		if rand.Float64() < 1-math.Pow(d1/september.interferenceRange, 6) {
			september.buckets[i].In(september.nanosecByPacket(size))
		} else if rand.Float64() < 1-math.Pow(d2/september.interferenceRange, 6) {
			september.buckets[i].In(september.ackIntererence())
		}
	}

	// The packet takes the adventure in the air (fading, etc.)
	if rand.Float64() > september.deliverRate(destination, dist) {
		return false
	}

	// Go through destination bucket
	if !september.buckets[destination].In(september.nanosecByPacket(size)) {
		return false
	}

	// The packet is gonna be delivered!
	return true
}

func (september *september2nd) SendBroadcast(source int, size int, underlying []int) []int {
	if !september.positionManager.IsEnabled(source) {
		return underlying[:0]
	}
	// Go through source bucket
	if !september.buckets[source].In(size) {
		return underlying[:0]
	}

	count := 0
	for _, i := range september.positionManager.Enabled() {
		dist := september.positionManager.Distance(source, i)
		if dist < september.interferenceRange {
			if !september.buckets[i].In(september.nanosecByPacket(size)) {
				// Go through destination bucket. If rejected by the bucket, the
				// broadcasted packet should not be delivered to this node
				continue
			}
		}

		// The packet takes the adventure in the air (fading, etc.). There's still
		// a possibility the packet is not delivered to this node
		if rand.Float64() > september.deliverRate(i, dist) {
			continue
		}

		// The packet is gonna be delivered!
		underlying[count] = i
		count++
	}
	return underlying[:count]
}
