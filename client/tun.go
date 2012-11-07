package client

import (
	"io"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	cIFF_TUN   = 0x0001
	cIFF_NO_PI = 0x1000
)

const (
	cMAX_PACKET_SIZE = 1500 //MTU
)

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

type tunIF struct {
	name string
	file *os.File
}

func (tun *tunIF) Name() string {
	return tun.name
}

func (tun *tunIF) createInterface(ifPattern string) (err error) {
	var req ifReq
	req.Flags = cIFF_TUN | cIFF_NO_PI
	copy(req.Name[:], ifPattern)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, tun.file.Fd(), uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		return errno
	}
	tun.name = strings.Trim(string(req.Name[:]), "\x00")
	return
}

func newTun(ifPattern string) (tun *tunIF, err error) {
	file, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	tun = &tunIF{name: "", file: file}
	err = tun.createInterface(ifPattern)
	if err != nil {
		file.Close()
		return nil, err
	}
	return
}

// Read a packet from TUN device. 
func (tun *tunIF) Read() (packet []byte, err error) {
	buf := make([]byte, cMAX_PACKET_SIZE)
	n, err := tun.file.Read(buf)
	if err != nil && err != io.EOF {
		return
	}
	packet = buf[:n]
	return packet, nil
}

// Write a packet into TUN device (blocking)
func (tun *tunIF) Write(packet []byte) (err error) {
	_, err = tun.file.Write(packet)
	return
}
