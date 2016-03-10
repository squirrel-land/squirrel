package main

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/squirrel-land/squirrel"
	"github.com/squirrel-land/squirrel/common"
	"github.com/squirrel-land/water/waterutil"
)

type client struct {
	Link *common.Link
	Addr net.HardwareAddr
}

type Master struct {
	addressPool     *addressPool
	clients         []*client
	addrReverse     *addressReverse
	positionManager squirrel.PositionManager

	mobilityManager squirrel.MobilityManager
	september       squirrel.September
}

func NewMaster(network *net.IPNet, mobilityManager squirrel.MobilityManager, september squirrel.September) (master *Master) {
	master = &Master{addressPool: newAddressPool(network), addrReverse: newAddressReverse(), mobilityManager: mobilityManager, september: september}
	master.clients = make([]*client, master.addressPool.Capacity()+1, master.addressPool.Capacity()+1)
	master.positionManager = NewPositionManager(master.addressPool.Capacity()+1, master.addrReverse)
	master.mobilityManager.Initialize(master.positionManager)
	master.september.Initialize(master.positionManager)
	return
}

func (master *Master) clientJoin(identity int, addr net.HardwareAddr, link *common.Link) {
	master.clients[identity] = &client{Link: link, Addr: addr}
	master.positionManager.Enable(identity)
	master.addrReverse.Add(addr, identity)
	ipAddr, _ := master.addressPool.GetAddress(identity)
	fmt.Printf("%v joined\n", ipAddr)
}

func (master *Master) clientLeave(identity int) {
	master.addrReverse.Remove(master.clients[identity].Addr)
	master.clients[identity] = nil
	master.positionManager.Disable(identity)
	addr, _ := master.addressPool.GetAddress(identity)
	fmt.Printf("%v left\n", addr)
}

func (master *Master) accept(listener net.Listener) (identity int, err error) {
	var connection net.Conn
	connection, err = listener.Accept()
	if err != nil {
		return
	}
	link := common.NewLink(connection)

	var req *common.JoinReq
	req, err = link.GetJoinReq()
	if err != nil {
		return
	}

	for identity = 1; identity < len(master.clients); identity++ {
		if master.clients[identity] == nil {
			break
		}
	}
	if identity == len(master.clients) {
		err = errors.New("Adress poll is full")
		link.SendJoinRsp(&common.JoinRsp{Error: err})
		return
	}

	var addr net.IP
	addr, err = master.addressPool.GetAddress(identity)
	if err != nil {
		return
	}
	err = link.SendJoinRsp(&common.JoinRsp{Address: addr, Mask: master.addressPool.Network.Mask, Error: nil})
	if err != nil {
		return
	}
	master.clientJoin(identity, req.MACAddr, link)
	link.StartRoutines()
	return identity, nil
}

func (master *Master) frameHandler(myIdentity int) {
	var (
		bufferedFrame *common.BufferedFrame
		wg            = new(sync.WaitGroup)
		underlying    = make([]int, master.addressPool.Capacity()+1)
	)

	for {
		bufferedFrame = <-master.clients[myIdentity].Link.ReadFrame
		if master.clients[myIdentity].Link.Error != nil {
			master.clientLeave(myIdentity)
			if bufferedFrame != nil {
				bufferedFrame.Return()
			}
			return
		}
		dst := waterutil.MACDestination(bufferedFrame.Frame)
		if waterutil.IsBroadcast(dst) || waterutil.IsIPv4Multicast(dst) {
			recipients := master.september.SendBroadcast(myIdentity, len(bufferedFrame.Frame), underlying)
			for _, id := range recipients {
				if master.clients[id] != nil { // This is mostly not necessary. Added due to not using locks, just in case.
					wg.Add(1)
					master.clients[id].Link.WriteWithNotify(bufferedFrame, wg)
				}
			}
			wg.Wait()
			bufferedFrame.Return()
		} else { // unicast
			dstId, ok := master.addrReverse.Get(waterutil.MACDestination(bufferedFrame.Frame))
			if ok && master.september.SendUnicast(myIdentity, dstId, len(bufferedFrame.Frame)) {
				master.clients[dstId].Link.WriteAndReturnBuffer(bufferedFrame)
			} else {
				bufferedFrame.Return()
			}
		}
	}
}

func (master *Master) Run(laddr string) (err error) {
	var (
		listener net.Listener
		identity int
	)

	listener, err = net.Listen("tcp", laddr)
	if err != nil {
		return
	}
	for {
		identity, err = master.accept(listener)
		if err != nil {
			continue
		}
		go master.frameHandler(identity)
	}
	return
}
