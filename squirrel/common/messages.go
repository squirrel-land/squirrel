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

type MessageType struct {
    Type        uint8       //256 kinds should be enough for all messages, huh?
}

type JoinReq struct {
    Identity    uint32
}

type JoinRsp struct {
    Address     net.IP
    Mask        net.IP
}

type Packet struct {
    Length      uint16
    NextHop		net.IP
    Packet		[]byte
}
