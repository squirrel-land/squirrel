package common

import (
	"encoding/gob"
	"net"
)

type Link struct {
	connection  *net.IPConn
	buffer      chan *BufferedPacket // buffer pool. It owns instances of BufferedPacket
	ReadPacket  chan *BufferedPacket // The channel used to read a packet from this Link
	WritePacket chan *Packet // The channel to write a packet into this Link
	encoder     *gob.Encoder
	decoder     *gob.Decoder
}

func NewLink(conn *net.IPConn) (link *Link) {
	return &Link{
		connection:  conn,
		buffer:      make(chan *BufferedPacket, BUFFERSIZE),
		ReadPacket:  make(chan *BufferedPacket),
		WritePacket: make(chan *Packet, BUFFERSIZE),
		encoder:     gob.NewEncoder(conn),
		decoder:     gob.NewDecoder(conn),
	}
}

// Send a JoinReq to the Link. Blocking.
func (link *Link) SendJoinReq(req *JoinReq) (err error) {
	return link.encoder.Encode(req)
}

// Get a JoinReq from the Link. Blocking.
func (link *Link) GetJoinReq() (req *JoinReq, err error) {
	req = new(JoinReq)
	err = link.decoder.Decode(req)
	return
}

// Send a JoinRsp to the Link. Blocking.
func (link *Link) SendJoinRsp(rsp *JoinRsp) (err error) {
	return link.encoder.Encode(rsp)
}

// Get a JoinRsp from the Link. Blocking.
func (link *Link) GetJoinRsp() (rsp *JoinRsp, err error) {
	rsp = new(JoinRsp)
	err = link.decoder.Decode(rsp)
	return
}

// Start routines that handle non-blocking read/write. This should be called only after initialization(req/rsp) process.
func (link *Link) StartRoutines() {
	for i := 0; i < BUFFERSIZE; i++ {
		link.buffer <- NewBufferedPacket(link.buffer)
	}
	go link.readRoutine()
	go link.writeRoutine()
}

func (link *Link) readRoutine() {
	var t MessageType
	var buf *BufferedPacket
	for {
		link.decoder.Decode(&t)

		if t.Type == MSGPACKET {
			buf = <-link.buffer
			link.decoder.Decode(buf.Packet)
			link.ReadPacket <- buf
		}
	}
}

func (link *Link) writeRoutine() {
	var p *Packet
	for {
		p = <-link.WritePacket
		link.encoder.Encode(p)
	}
}
