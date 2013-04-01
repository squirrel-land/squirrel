package common

import (
	"encoding/gob"
	"errors"
	"net"
	"sync"
)

var (
	ConnectionClosed          = errors.New("Connection is closed.")
	UnKnownTypeInWriteChannel = errors.New("Unknown type sent to write channel.")
)

type notifiableBufferedFrame struct {
	BufferedFrame *BufferedFrame
	waitGroup     *sync.WaitGroup
}

// A Link can send or receive frames. It uses channels internally and is thread-safe. To avoid too much GC, a circular buffer is used for reading frames. Buffered frames are owned by Link and should be returned once they are not useful to whoever reads them from here.
type Link struct {
	Error      error // indicate whether there's any error encountered.
	connection net.Conn
	buffer     chan *BufferedFrame // buffer pool. It owns instances of BufferedFrame
	ReadFrame  chan *BufferedFrame // The channel used to read a frame from this Link. It is necessary to call .Return() when finishing using the BufferedFrame.
	writeFrame chan interface{}    // The channel to write a frame into this Link.
	encoder    *gob.Encoder
	decoder    *gob.Decoder
}

func NewLink(conn net.Conn) (link *Link) {
	return &Link{
		connection: conn,
		buffer:     make(chan *BufferedFrame, BUFFERSIZE),
		ReadFrame:  make(chan *BufferedFrame, BUFFERSIZE),
		writeFrame: make(chan interface{}, BUFFERSIZE),
		encoder:    gob.NewEncoder(conn),
		decoder:    gob.NewDecoder(conn),
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
		link.buffer <- NewBufferedFrame(link.buffer)
	}
	go link.readRoutine()
	go link.writeRoutine()
}

func (link *Link) readRoutine() {
	var t MessageType
	var buf *BufferedFrame
	for {
		if link.Error != nil {
			link.ReadFrame <- nil
			return
		}
		link.Error = link.decoder.Decode(&t)
		if t.Type == 0 {
			link.Error = ConnectionClosed
		}
		if link.Error != nil {
			link.ReadFrame <- nil
			return
		}
		if t.Type == MSGFRAME {
			buf = <-link.buffer
			link.Error = link.decoder.Decode(&buf.Frame)
			if link.Error != nil {
				link.ReadFrame <- buf
				return
			}
			link.ReadFrame <- buf
		}
	}
}

func (link *Link) writeRoutine() {
	var bufi interface{}
	messageType := MessageType{Type: MSGFRAME}
	for {
		if link.Error == nil {
			bufi = <-link.writeFrame
			link.Error = link.encoder.Encode(messageType)
			switch buf := bufi.(type) {
			case *BufferedFrame:
				if link.Error == nil {
					link.Error = link.encoder.Encode(buf.Frame)
				}
				buf.Return()
			case notifiableBufferedFrame:
				if link.Error == nil {
					link.Error = link.encoder.Encode(buf.BufferedFrame.Frame)
				}
				buf.waitGroup.Done()
			case Frame:
				if link.Error == nil {
					link.Error = link.encoder.Encode(buf)
				}
			default:
				link.Error = UnKnownTypeInWriteChannel
			}
		}
	}
}

// Write a buffered frame into the link.
// The buffered frame is returned to its owner as soon as it's sent completely to the network.
func (link *Link) WriteAndReturnBuffer(bufferedFrame *BufferedFrame) {
	link.writeFrame <- bufferedFrame
}

// Write a buffered frame into the link.
// A value is sent to notify as soon as the buffered frame is sent completely to the network. The caller needs to ensure that the buffered frame is returned (.Return()) after finishing using it.
func (link *Link) WriteWithNotify(bufferedFrame *BufferedFrame, wg *sync.WaitGroup) {
	nbuf := notifiableBufferedFrame{BufferedFrame: bufferedFrame, waitGroup: wg}
	link.writeFrame <- nbuf
}

// Write a frame into the link.
// Should only be used for unbuffered frame.
func (link *Link) WriteUnbuffered(frame Frame) {
	link.writeFrame <- frame
}
