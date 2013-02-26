package models

import (
	"github.com/songgao/squirrel/models/common"

	"math"
	"math/rand"
)

type september2nd struct {
	noDeliveryDistance float64
	interferenceRange  float64

	nodes    []*common.Position
	buckets  []*leakyBucket // measured by number of nanoseconds used;
	dataRate float64        // bit data rates in bit/nanosecond 

	// MAC layer time in nanoseconds
	difs  int
	sifs  int
	slot  int
	cwMin int

	// MAC layer frame properties in bits
	macFrameMaxBody  int
	macFrameOverhead int
}

func NewSeptember2nd() common.September {
	ret := new(september2nd)
	ret.dataRate = 54 * 1024 * 1024 * 1e-9 // 54 Mbps
	ret.slot = 9e3                         // 9 microseconds
	ret.sifs = 10e3                        // 10 microseconds
	ret.difs = ret.sifs + 2*ret.slot       // 28 microseconds
	ret.cwMin = 31
	ret.macFrameMaxBody = 2312 * 8
	ret.macFrameOverhead = 34
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

func (september *september2nd) Initialize(nodes []*common.Position) {
	september.nodes = nodes
	september.buckets = make([]*leakyBucket, len(nodes), len(nodes))
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
	return september.slot * september.cwMin / 2
}

func (september *september2nd) nanosecByPacket(packetSize int) int {
	return september.difs + september.cw() + september.nanosecByData(packetSize) + september.sifs + september.nanosecByData(0) // the last nanosecByData(0) is for MAC layer ACK
}

func (september *september2nd) ackIntererence() int {
	return september.difs + september.cw() + september.sifs + september.nanosecByData(0)
}

func (september *september2nd) SendUnicast(source int, destination int, size int) bool {
	p1 := september.nodes[source]
	p2 := september.nodes[destination]
	if p1 == nil || p2 == nil {
		return false
	}

	// Go through source bucket
	if !september.buckets[source].In(september.nanosecByPacket(size)) {
		return false
	}

	p1.Mu.RLock()
	p2.Mu.RLock()
	defer p1.Mu.RUnlock()
	defer p2.Mu.RUnlock()

	dist := distance(p1, p2)

	// Since the packet is out in the air, interference should be put on neighbor nodes
	for i := 1; i < len(september.nodes); i++ {
		n := september.nodes[i]
		if i != source && i != destination && n != nil {
			n.Mu.RLock()
			defer n.Mu.RUnlock()
			if rand.Float64() < 1-math.Pow(distance(p1, n)/september.interferenceRange, 5) {
				september.buckets[i].In(september.nanosecByPacket(size))
			} else if rand.Float64() < 1-math.Pow(distance(p2, n)/september.interferenceRange, 5) {
				september.buckets[i].In(september.ackIntererence())
			}
		}
	}

	// The packet takes the adventure in the air (fading, etc.)
	if dist > september.noDeliveryDistance*0.85 {
		if rand.Float64() < math.Pow(dist/september.noDeliveryDistance, 3) {
			return false
		}
	} else {
		if rand.Float64() < math.Pow(dist/september.noDeliveryDistance, 8) {
			return false
		}
	}

	// Go through destination bucket
	if !september.buckets[destination].In(september.nanosecByPacket(size)) {
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

	p1.Mu.RLock()
	defer p1.Mu.RUnlock()

	count := 0
	for i := 1; i < len(september.nodes); i++ {
		p2 := september.nodes[i]
		if p2 == nil {
			continue
		}

		p2.Mu.RLock()
		defer p2.Mu.RUnlock()

		dist := distance(p1, p2)
		if dist < september.interferenceRange {
			if !september.buckets[i].In(september.nanosecByPacket(size)) {
				// Go through destination bucket. If rejected by the bucket, the broadcasted packet should not be delivered to this node
				continue
			}
		}

		// The packet takes the adventure in the air (fading, etc.). There's still a possibility the packet is not delivered to this node
		if dist > september.noDeliveryDistance*0.85 {
			if rand.Float64() < math.Pow(dist/september.noDeliveryDistance, 3) {
				continue
			}
		} else {
			if rand.Float64() < math.Pow(dist/september.noDeliveryDistance, 8) {
				continue
			}
		}

		// The packet is gonna be delivered!
		underlying[count] = i
		count++
	}
	return underlying[:count]
}
