package main

import (
	"errors"
	"fmt"
	"net"
	"os/exec"

	"github.com/squirrel-land/squirrel/common"
	"github.com/squirrel-land/water"
)

type Client struct {
	link         *common.Link
	tap          *water.Interface
	routinesQuit chan error
}

// Create a new client along with a TAP network interface whose name is tapName
func NewClient(tapName string) (client *Client, err error) {
	tap, err := water.NewTAP(tapName)
	if err != nil {
		return nil, err
	}
	client = &Client{
		link:         nil,
		tap:          tap,
		routinesQuit: make(chan error),
	}
	return
}

func (client *Client) configureTap(joinRsp *common.JoinRsp) (err error) {
	m, _ := joinRsp.Mask.Size()
	addr := fmt.Sprintf("%s/%d", joinRsp.Address.String(), m)
	err = exec.Command("ip", "addr", "add", addr, "dev", client.tap.Name()).Run()
	if err != nil {
		return
	}
	err = exec.Command("ip", "link", "set", "dev", client.tap.Name(), "up").Run()
	return
}

func (client *Client) connect(masterAddr string, identity int) (err error) {
	connection, err := net.Dial("tcp", masterAddr)
	if err != nil {
		return
	}
	client.link = common.NewLink(connection)

	var ifce *net.Interface
	ifce, err = net.InterfaceByName(client.tap.Name())
	err = client.link.SendJoinReq(&common.JoinReq{Identity: identity, MACAddr: ifce.HardwareAddr})
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
	err = client.configureTap(rsp)
	if err != nil {
		return
	}
	client.link.StartRoutines()
	return
}

func (client *Client) tap2master() {
	var err error
	buffer := make(chan *common.BufferedFrame, common.BUFFERSIZE)
	for i := 0; i < common.BUFFERSIZE; i++ {
		buffer <- common.NewBufferedFrame(buffer)
	}
	for {
		buf := <-buffer
		_, err = client.tap.Read(buf.Frame)
		if err != nil {
			client.routinesQuit <- err
			return
		}
		client.link.WriteAndReturnBuffer(buf)
	}
}

func (client *Client) master2tap() {
	var buf *common.BufferedFrame
	var err error
	for {
		buf = <-client.link.ReadFrame
		if client.link.Error != nil {
			client.routinesQuit <- client.link.Error
			return
		}
		_, err = client.tap.Write(buf.Frame)
		buf.Return()
		if err != nil {
			client.routinesQuit <- err
			return
		}
	}
}

// Run the client, and block until all routines exit or any error is ecountered.
// It connects to a master with address masterAddr, proceeds with JoinReq/JoinRsp process, configures the TAP device, and at last, start routines that carry MAC frames back and forth between the TAP device and the master.
// masterAddr: should be host:port format where host can be either IP address or hostname/domainName.
// identity: is an integer that is unique among all clients to identify different clients. Master uses identity to assign IP address to each client.
func (client *Client) Run(masterAddr string, identity int) (err error) {
	err = client.connect(masterAddr, identity)
	if err != nil {
		return
	}

	go client.tap2master()
	go client.master2tap()

	err = <-client.routinesQuit // first finished or error routine
	if err != nil {
		return
	}
	err = <-client.routinesQuit // second finished or error routine
	return
}
