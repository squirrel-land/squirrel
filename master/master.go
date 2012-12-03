package master

import (
	"../common"
	"errors"
	"fmt"
	"net"
	"sync"
)

type Master struct {
	addressPool     *addressPool
	clients         []*common.Link
	mobileNodes     []*Position
	mobilityManager MobilityManager
	september       September
}

func NewMaster(network *net.IPNet, mobilityManager MobilityManager, september September) (master *Master) {
	master = &Master{addressPool: newAddressPool(network), mobilityManager: mobilityManager, september: september}
	master.clients = make([]*common.Link, master.addressPool.Capacity()+1, master.addressPool.Capacity()+1)
	master.mobileNodes = make([]*Position, master.addressPool.Capacity()+1, master.addressPool.Capacity()+1)
	master.mobilityManager.Initialize(master.mobileNodes)
	master.september.Initialize(master.mobileNodes)
	return
}

func (master *Master) clientJoin(identity int, link *common.Link) {
	master.clients[identity] = link
	master.mobileNodes[identity] = master.mobilityManager.GenerateNewNode()
	addr, _ := master.addressPool.GetAddress(identity)
	fmt.Printf("%v joined\n", addr)
}

func (master *Master) clientLeave(identity int) {
	master.clients[identity] = nil
	master.mobileNodes[identity] = nil
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
	master.clientJoin(req.Identity, link)
	link.StartRoutines()
	return req.Identity, nil
}

func (master *Master) packetHandler(myIdentity int) {
	var (
		bufferedPacket *common.BufferedPacket
		nextHopId      int
		err            error
		wg             = new(sync.WaitGroup)
		underlying     = make([]int, master.addressPool.Capacity()+1, master.addressPool.Capacity()+1)
	)

	for {
		bufferedPacket = <-master.clients[myIdentity].ReadPacket
		if master.clients[myIdentity].Error != nil {
			master.clientLeave(myIdentity)
			if bufferedPacket != nil {
				bufferedPacket.Return()
			}
			return
		}
		if master.addressPool.IsBroadcast(bufferedPacket.Packet.NextHop) || bufferedPacket.Packet.NextHop.IsMulticast() {
			recipients := master.september.SendBroadcast(myIdentity, underlying)
			for _, id := range recipients {
				if master.clients[id] != nil { // This is mostly not necessary. Added due to not using locks, just in case.
					wg.Add(1)
					master.clients[id].WriteWithNotify(bufferedPacket, wg)
				}
			}
			wg.Wait()
			bufferedPacket.Return()
		} else { // unicast
			nextHopId, err = master.addressPool.GetIdentity(bufferedPacket.Packet.NextHop)
			if err == nil && master.clients[nextHopId] != nil && master.september.SendUnicast(myIdentity, nextHopId) {
				master.clients[nextHopId].WriteAndReturnBuffer(bufferedPacket)
			} else {
				bufferedPacket.Return()
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
		go master.packetHandler(identity)
	}
	return
}
