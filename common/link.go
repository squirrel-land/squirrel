package common

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
)

// A Link can send or receive frames. It uses channels internally and is
// thread-safe.
type Link struct {
	connection net.Conn
	encoder    *gob.Encoder
	decoder    *gob.Decoder

	incoming      chan *ReusableSlice
	outgoing      chan *ReusableSlice
	incomingError error
}

func (l *Link) ReadFrame() (frame *ReusableSlice, ok bool) {
	frame, ok = <-l.incoming
	return
}

func (l *Link) WriteFrame(frame *ReusableSlice) {
	l.outgoing <- frame
}

func (l *Link) Done() {
	close(l.outgoing)
}

// IncomingError returns the error (if any) happened while decoding an incoming
// message.  Note: if there's an error in encoding outgoing messages, it is
// considered an implementation and log.Fatalf is called.
func (l *Link) IncomingError() error {
	return l.incomingError
}

func NewLink(conn net.Conn) (link *Link) {
	return &Link{
		connection: conn,
		encoder:    gob.NewEncoder(conn),
		decoder:    gob.NewDecoder(conn),
		incoming:   make(chan *ReusableSlice, 64),
		outgoing:   make(chan *ReusableSlice, 64),
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
	go link.readRoutine()
	go link.writeRoutine()
}

func (link *Link) failIncoming(err error) {
	link.incomingError = err
	close(link.incoming)
}

func (link *Link) readRoutine() {
	pool := NewSlicePool(1522)
	var (
		t   MsgType
		buf *ReusableSlice
	)
	var err error
	for {
		if err = link.decoder.Decode(&t); err != nil {
			if err != io.EOF {
				link.failIncoming(fmt.Errorf("decoding MsgType error: %v", err))
			} else {
				link.failIncoming(nil)
			}
			return
		}
		if t == MSGFRAME {
			buf = pool.Get()
			if err = link.decoder.Decode(buf.SlicePtr()); err != nil {
				link.failIncoming(fmt.Errorf("decoding frame error: %v", err))
				return
			}
			link.incoming <- buf
		} else {
			link.failIncoming(fmt.Errorf("unexpected MsgType: %d", t))
			return
		}
	}
}

func (link *Link) writeRoutine() {
	var err error
	for buf := range link.outgoing {
		if err = link.encoder.Encode(MSGFRAME); err != nil {
			log.Fatalf("error encoding MSGFRAME: %v\n", err)
		}
		if err = link.encoder.Encode(buf.Slice()); err != nil {
			log.Fatalf("error encoding MSGFRAME: %v\n", err)
		}
		buf.Done()
	}
}
