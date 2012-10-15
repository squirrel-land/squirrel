package common

import (
	"encoding/gob"
	"net"
)

type Link struct {
	connection  *net.IPConn
	buffer      chan *BufferedPacket
	ReadPacket  chan *BufferedPacket
	WritePacket chan *BufferedPacket
	encoder     *gob.Encoder
	decoder     *gob.Decoder
}

func NewLink(conn *net.IPConn) (link *Link) {
	return &Link{
		connection:  conn,
		buffer:      make(chan *BufferedPacket, BUFFERSIZE),
		ReadPacket:  make(chan *BufferedPacket),
		WritePacket: make(chan *BufferedPacket),
		encoder:     gob.NewEncoder(conn),
		decoder:     gob.NewDecoder(conn),
	}
}

func (link *Link) SendJoinReq(req *JoinReq) (err error) {
	return link.encoder.Encode(req)
}

func (link *Link) GetJoinReq() (req *JoinReq, err error) {
	req = new(JoinReq)
	err = link.decoder.Decode(req)
	return
}

func (link *Link) SendJoinRsp(rsp *JoinRsp) (err error) {
	return link.encoder.Encode(rsp)
}

func (link *Link) GetJoinRsp() (rsp *JoinRsp, err error) {
	rsp = new(JoinRsp)
	err = link.decoder.Decode(rsp)
	return
}

// should be called only after initialization(req/rsp)
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
	var buf *BufferedPacket
	for {
		buf = <-link.WritePacket
		link.encoder.Encode(buf.Packet)
		buf.Return()
	}
}
