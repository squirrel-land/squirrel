package common

const (
	BUFFERSIZE = 32
)

type BufferedPacket struct {
	Packet *Packet
	owner  chan *BufferedPacket
}

func NewBufferedPacket(owner chan *BufferedPacket) *BufferedPacket {
	packet := new(Packet)
	return &BufferedPacket{Packet: packet, owner: owner}
}

func (buf *BufferedPacket) Return() {
	buf.owner <- buf
}
