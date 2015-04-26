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
	addrReverse     map[string]int
	positionManager squirrel.PositionManager
	mu              sync.RWMutex // just for addrReverse, since maps are not thread-safe.
	mobilityManager squirrel.MobilityManager
	september       squirrel.September
}

func NewMaster(network *net.IPNet, mobilityManager squirrel.MobilityManager, september squirrel.September) (master *Master) {
	master = &Master{addressPool: newAddressPool(network), addrReverse: make(map[string]int), mobilityManager: mobilityManager, september: september}
	master.clients = make([]*client, master.addressPool.Capacity()+1, master.addressPool.Capacity()+1)
	master.positionManager = NewPositionManager(master.addressPool.Capacity() + 1)
	master.mobilityManager.Initialize(master.positionManager)
	master.september.Initialize(master.positionManager)
	return
}

func (master *Master) clientJoin(identity int, addr net.HardwareAddr, link *common.Link) {
	master.mu.Lock()
	defer master.mu.Unlock()
	master.clients[identity] = &client{Link: link, Addr: addr}
	master.positionManager.Enable(identity)
	master.addrReverse[addr.String()] = identity
	ipAddr, _ := master.addressPool.GetAddress(identity)
	fmt.Printf("%v joined\n", ipAddr)
}

func (master *Master) clientLeave(identity int) {
	master.mu.Lock()
	defer master.mu.Unlock()
	delete(master.addrReverse, master.clients[identity].Addr.String())
	master.clients[identity] = nil
	master.positionManager.Disable(identity)
	addr, _ := master.addressPool.GetAddress(identity)
	fmt.Printf("%v left\n", addr)
}

func (master *Master) accept(listener net.Listener) (identity int, err error) {
	connection, err := listener.Accept()
	if err != nil {
		return
	}
	link := common.NewLink(connection)

	req, err := link.GetJoinReq()
	if err != nil {
		return
	}
	if master.clients[req.Identity] != nil {
		link.SendJoinRsp(&common.JoinRsp{Success: false})
		err = errors.New("Duplicate identity")
		return
	}
	addr, err := master.addressPool.GetAddress(req.Identity)
	if err != nil {
		return
	}
	err = link.SendJoinRsp(&common.JoinRsp{Address: addr, Mask: master.addressPool.Network.Mask, Success: true})
	if err != nil {
		return
	}
	master.clientJoin(req.Identity, req.MACAddr, link)
	link.StartRoutines()
	return req.Identity, nil
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
			master.mu.RLock() // maps are not thread-safe
			dstId, ok := master.addrReverse[waterutil.MACDestination(bufferedFrame.Frame).String()]
			client := master.clients[dstId]
			master.mu.RUnlock()
			if ok && master.september.SendUnicast(myIdentity, dstId, len(bufferedFrame.Frame)) {
				client.Link.WriteAndReturnBuffer(bufferedFrame)
			} else {
				bufferedFrame.Return()
			}
		}
	}
}

func (master *Master) Run(laddr string) (err error) {
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		return
	}
	for {
		identity, err := master.accept(listener)
		if err != nil {
			continue
		}
		go master.frameHandler(identity)
	}
	return
}
