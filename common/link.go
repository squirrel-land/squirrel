package common

import (
	"encoding/gob"
	"errors"
	"net"
)

var (
	ConnectionClosed          = errors.New("Connection is closed.")
	UnKnownTypeInWriteChannel = errors.New("Unknown type sent to write channel.")
)

type notifiableBufferedPacket struct {
	BufferedPacket *BufferedPacket
	notifyChan     chan byte
}

func (this *notifiableBufferedPacket) notify() {
	this.notifyChan <- 0
}

type Link struct {
	Error       error // indicate whether there's any error encountered.
	connection  net.Conn
	buffer      chan *BufferedPacket // buffer pool. It owns instances of BufferedPacket
	ReadPacket  chan *BufferedPacket // The channel used to read a packet from this Link. It is necessary to call .Return() when finishing using the BufferedPacket.
	writePacket chan interface{}     // The channel to write a packet into this Link.
	encoder     *gob.Encoder
	decoder     *gob.Decoder
}

func NewLink(conn net.Conn) (link *Link) {
	return &Link{
		connection:  conn,
		buffer:      make(chan *BufferedPacket, BUFFERSIZE),
		ReadPacket:  make(chan *BufferedPacket, BUFFERSIZE),
		writePacket: make(chan interface{}, BUFFERSIZE),
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
		if link.Error != nil {
			link.ReadPacket <- nil
			return
		}
		link.Error = link.decoder.Decode(&t)
		if t.Type == 0 {
			link.Error = ConnectionClosed
		}
		if link.Error != nil {
			link.ReadPacket <- nil
			return
		}
		if t.Type == MSGPACKET {
			buf = <-link.buffer
			link.Error = link.decoder.Decode(buf.Packet)
			if link.Error != nil {
				link.ReadPacket <- buf
				return
			}
			link.ReadPacket <- buf
		}
	}
}

func (link *Link) writeRoutine() {
	var bufi interface{}
	packetType := MessageType{Type: MSGPACKET}
	for {
		if link.Error == nil {
			bufi = <-link.writePacket
			link.Error = link.encoder.Encode(packetType)
			switch buf := bufi.(type) {
			case *BufferedPacket:
				if link.Error == nil {
					link.Error = link.encoder.Encode(buf.Packet)
				}
				buf.Return()
			case notifiableBufferedPacket:
				if link.Error == nil {
					link.Error = link.encoder.Encode(buf.BufferedPacket.Packet)
				}
				buf.notify()
			case *Packet:
				if link.Error == nil {
					link.Error = link.encoder.Encode(buf)
				}
			default:
				link.Error = UnKnownTypeInWriteChannel
			}
		}
	}
}

// Write a buffered packet into the link.
// The buffered packet is returned to its owner as soon as it's sent completely to the network.
func (link *Link) WriteAndReturnBuffer(bufferedPacket *BufferedPacket) {
	link.writePacket <- bufferedPacket
}

// Write a buffered packet into the link.
// A value is sent to notify as soon as the buffered packet is sent completely to the network. The caller needs to ensure that the buffered packet is returned (.Return()) after finishing using it.
func (link *Link) WriteWithNotify(bufferedPacket *BufferedPacket, notify chan byte) {
	nbuf := notifiableBufferedPacket{BufferedPacket: bufferedPacket, notifyChan: notify}
	link.writePacket <- nbuf
}

// Write a packet into the link.
// Should only be used for unbuffered packet.
func (link *Link) WriteUnbuffered(packet *Packet) {
	link.writePacket <- packet
}
