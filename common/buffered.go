package common

const (
	BUFFERSIZE = 32
	FRAMESIZE  = 1522
)

// a Frame wrapper that holds a reference to the channel that owns it
type BufferedFrame struct {
	Frame Frame
	owner chan *BufferedFrame
}

func NewBufferedFrame(owner chan *BufferedFrame) *BufferedFrame {
	frame := make(Frame, FRAMESIZE)
	return &BufferedFrame{Frame: frame, owner: owner}
}

func (buf *BufferedFrame) Resize(length int) {
	buf.Frame = buf.Frame[:length]
}

// send the BufferedFrame back to its owner channel for further use
func (buf *BufferedFrame) Return() {
	if buf.owner != nil {
		buf.Resize(FRAMESIZE)
		buf.owner <- buf
	}
}
