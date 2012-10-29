package master

import (
	"../common"
	"fmt"
	"net"
)

type Master struct {
	addressPool *addressPool
	clients     []*common.Link
}

func NewMaster(network *net.IPNet) (master *Master) {
	master = &Master{addressPool: newAddressPool(network)}
	master.clients = make([]*common.Link, master.addressPool.Capacity()+1, master.addressPool.Capacity()+1)
	return
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
	addr, err := master.addressPool.GetAddress(req.Identity)
	if err != nil {
		return
	}
	err = link.SendJoinRsp(&common.JoinRsp{Address: addr, Mask: master.addressPool.Network.Mask})
	if err != nil {
		return
	}
	master.clients[req.Identity] = link
	link.StartRoutines()
	fmt.Printf("%v (%v) joined\n", addr, connection.(*net.TCPConn).RemoteAddr())
	return req.Identity, nil
}

func (master *Master) packetHandler(myIdentity int) {
	var (
		bufferedPacket         *common.BufferedPacket
		notifyCount, nextHopId int
		err                    error
	)
	notify := make(chan byte)
	for {
		bufferedPacket = <-master.clients[myIdentity].ReadPacket
		if master.clients[myIdentity].Error != nil {
			addr, _ := master.addressPool.GetAddress(myIdentity)
			fmt.Printf("%v left\n", addr)
			master.clients[myIdentity] = nil
			return
		}
		if master.addressPool.IsBroadcast(bufferedPacket.Packet.NextHop) {
			notifyCount = 0
			for i := 1; i < len(master.clients); i++ {
				if master.clients[i] != nil {
					master.clients[i].Write(bufferedPacket, notify)
					notifyCount = notifyCount + 1
				}
			}
			for i := 0; i < notifyCount; i++ {
				<-notify
			}
		} else {
			nextHopId, err = master.addressPool.GetIdentity(bufferedPacket.Packet.NextHop)
			if err == nil && master.clients[nextHopId] != nil {
				master.clients[nextHopId].Write(bufferedPacket, nil)
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
