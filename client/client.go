package client

import (
	"github.com/songgao/squirrel/common"
	"errors"
	"fmt"
	"net"
	"os/exec"
)

type Client struct {
	link         *common.Link
	tun          *tunIF
	routes       *routes
	routinesQuit chan error
}

// Create a new client along with a TUN network interface whose name is tunName
func NewClient(tunName string) (client *Client, err error) {
	tun, err := newTun(tunName)
	client = &Client{
		link:         nil,
		tun:          tun,
		routinesQuit: make(chan error),
	}
	return
}

func (client *Client) configureTun(joinRsp *common.JoinRsp) (err error) {
	m, _ := joinRsp.Mask.Size()
	addr := fmt.Sprintf("%s/%d", joinRsp.Address.String(), m)
	err = exec.Command("ip", "addr", "add", addr, "dev", client.tun.Name()).Run()
	if err != nil {
		return
	}
	err = exec.Command("ip", "link", "set", "dev", client.tun.Name(), "up").Run()
	return
}

func (client *Client) connect(masterAddr string, identity int) (err error) {
	connection, err := net.Dial("tcp", masterAddr)
	if err != nil {
		return
	}
	client.link = common.NewLink(connection)

	err = client.link.SendJoinReq(&common.JoinReq{Identity: identity})
	if err != nil {
		return
	}
	rsp, err := client.link.GetJoinRsp()
	if err != nil {
		return
	}
	if rsp.Success != true {
		return errors.New("Join failed. Possiblly the Identity number in config file is duplicate on master.")
	}
	err = client.configureTun(rsp)
	if err != nil {
		return
	}
	client.routes = newRoutes(client.tun.Name())
	client.link.StartRoutines()
	return
}

func (client *Client) tun2master() {
	var pac []byte
	var err error
	for {
		pac, err = client.tun.Read()
		if err != nil {
			client.routinesQuit <- err
			return
		}
		packet := &common.Packet{Packet: pac}
		packet.NextHop = client.routes.Route(packet.Destination())
		client.link.WriteUnbuffered(packet)
	}
}

func (client *Client) master2tun() {
	var buf *common.BufferedPacket
	var err error
	for {
		buf = <-client.link.ReadPacket
		if client.link.Error != nil {
			client.routinesQuit <- client.link.Error
			return
		}
		err = client.tun.Write(buf.Packet.Packet)
		buf.Return()
		if err != nil {
			client.routinesQuit <- err
			return
		}
	}
}

// Run the client, and block until all routines exit or any error is ecountered.
// It connects to a master with address masterAddr, proceeds with JoinReq/JoinRsp process, configures the TUN device, and at last, start routines that carry packets back and forth between the TUN device and the master.
// masterAddr: should be host:port format where host can be either IP address or hostname/domainName.
// identity: is an integer that is unique among all clients to identify different clients. Master uses identity to assign IP address to each client.
func (client *Client) Run(masterAddr string, identity int) (err error) {
	err = client.connect(masterAddr, identity)
	if err != nil {
		return
	}

	go client.tun2master()
	go client.master2tun()

	err = <-client.routinesQuit // first finished or error routine
	if err != nil {
		return
	}
	err = <-client.routinesQuit // second finished or error routine
	return
}
