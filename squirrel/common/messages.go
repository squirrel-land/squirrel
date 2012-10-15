package common

import (
	"net"
)

// Message types
const (
	MSGJOINREQ = iota
	MSGJOINRSP
	MSGPACKET
)

// sent before each message to help identify message type
type MessageType struct {
	Type uint8 //256 kinds should be enough for all messages, huh?
}

// sent from client to master, representing request to join
type JoinReq struct {
	Identity uint32
}

// sent from master back to client, indicating assigned IP address and Mask
type JoinRsp struct {
	Address net.IP
	Mask    net.IP
}

// represent a network layer packet
type Packet struct {
	Length  uint16
	NextHop net.IP
	Packet  []byte
}
