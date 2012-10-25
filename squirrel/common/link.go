package common

import (
	"encoding/gob"
	"net"
)

type notifiableBufferedPacket struct {
	BufferedPacket *BufferedPacket
	notify         chan byte
}

func (this *notifiableBufferedPacket) NotifyOrReturn() {
	if this.notify != nil {
		this.notify <- 0
	} else {
		this.BufferedPacket.Return()
	}
}

type Link struct {
	connection  net.Conn
	buffer      chan *BufferedPacket          // buffer pool. It owns instances of BufferedPacket
	ReadPacket  chan *BufferedPacket          // The channel used to read a packet from this Link. It is necessary to call .Return() when finishing using the BufferedPacket.
	writePacket chan notifiableBufferedPacket // The channel to write a packet into this Link.
	encoder     *gob.Encoder
	decoder     *gob.Decoder
}

func NewLink(conn net.Conn) (link *Link) {
	return &Link{
		connection:  conn,
		buffer:      make(chan *BufferedPacket, BUFFERSIZE),
		ReadPacket:  make(chan *BufferedPacket, BUFFERSIZE),
		writePacket: make(chan notifiableBufferedPacket, BUFFERSIZE),
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
	var nbuf notifiableBufferedPacket
	packetType := MessageType{Type: MSGPACKET}
	for {
		nbuf = <-link.writePacket
		link.encoder.Encode(packetType)
		link.encoder.Encode(nbuf.BufferedPacket.Packet)
		nbuf.NotifyOrReturn()
	}
}

// Write a buffered packet into the link.
// If notify is nil, the buffered packet is returned to its owner as soon as it's sent completely to the network; otherwise, a value is sent to notify as soon as the buffered packet is sent completely to the network.
// If notify is not nil, the caller needs to ensure that the buffered packet is returned (.Return()) after finishing using it.
func (link *Link) Write(bufferedPacket *BufferedPacket, notify chan byte) {
	nbuf := notifiableBufferedPacket{BufferedPacket: bufferedPacket, notify: notify}
	link.writePacket <- nbuf
}

// Write a packet into the link.
// Should only be used for unbuffered packet.
func (link *Link) WriteUnbuffered(packet *Packet) {
	link.writePacket <- notifiableBufferedPacket{BufferedPacket: &BufferedPacket{Packet: packet}}
}
