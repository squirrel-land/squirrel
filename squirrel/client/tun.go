package client

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

const (
	IFF_TUN   = 0x0001
	IFF_NO_PI = 0x1000
)

const (
	MAX_PACKET_SIZE = 1500 //MTU
)

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

type Tun struct {
	name string
	file *os.File
}

func (tun *Tun) Name() string {
	return tun.name
}

func (tun *Tun) createInterface(ifPattern string) (err error) {
	var req ifReq
	req.Flags = IFF_TUN | IFF_NO_PI
	copy(req.Name[:], ifPattern)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, tun.file.Fd(), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		return errno
	}
	tun.name = string(req.Name[:])
	return
}

func NewTun(ifPattern string) (tun *Tun, err error) {
	file, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	tun = &Tun{name: "", file: file}
	err = tun.createInterface(ifPattern)
	if err != nil {
		file.Close()
		return nil, err
	}
	return
}

// Read a packet from TUN device. 
func (tun *Tun) Read() (packet []byte, err error) {
	buf := make([]byte, MAX_PACKET_SIZE)
	n, err := tun.file.Read(buf)
	if err != nil && err != io.EOF {
		return
	}
	packet = buf[:n]
	return packet, nil
}

// Write a packet into TUN device (blocking)
func (tun *Tun) Write(packet []byte) (err error) {
	_, err = tun.file.Write(packet)
	return
}
