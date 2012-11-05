package common

import (
	"net"
)

// Message types
const (
	_ = iota
	MSGJOINREQ
	MSGJOINRSP
	MSGPACKET
)

// sent before each message to help identify message type
type MessageType struct {
	Type uint8 //256 kinds should be enough for all messages, huh?
}

// sent from client to master, representing request to join
type JoinReq struct {
	Identity int
}

// sent from master back to client, indicating assigned IP address and Mask
type JoinRsp struct {
	Address net.IP
	Mask    net.IPMask
	Success bool
}

// represent a network layer packet
type Packet struct {
	NextHop net.IP //since there's no MAC address involved, a nexthop IP address (but not MAC address) is placed here along with the actual packet. It's used by the server to determine which node to send this packet to, in order to support (multi-hop) routing.
	Packet  []byte
}

func (packet *Packet) Source() net.IP {
	return IPFrom(packet.Packet[12:16])
}

func (packet *Packet) Destination() net.IP {
	return IPFrom(packet.Packet[16:20])
}

func (packet *Packet) TTL() byte {
	return packet.Packet[8]
}
