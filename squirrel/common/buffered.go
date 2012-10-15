package common

const (
	BUFFERSIZE = 32
)

// a Packet wrapper that holds a reference to the channel that owns it
type BufferedPacket struct {
	Packet *Packet
	owner  chan *BufferedPacket
}

func NewBufferedPacket(owner chan *BufferedPacket) *BufferedPacket {
	packet := new(Packet)
	return &BufferedPacket{Packet: packet, owner: owner}
}

// send the BufferedPacket back to its owner channel for further use
func (buf *BufferedPacket) Return() {
	buf.owner <- buf
}
