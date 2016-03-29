package common

import (
	"net"
)

type MsgType uint8

// Message types
const (
	_ MsgType = iota
	MSGJOINREQ
	MSGJOINRSP
	MSGFRAME
)

// sent from client to master, representing request to join
type JoinReq struct {
	MACAddr net.HardwareAddr
}

// sent from master back to client, indicating assigned IP address and Mask
type JoinRsp struct {
	Address net.IP
	Mask    net.IPMask
	Error   error
}

// represent a MAC frame
type Frame []byte
