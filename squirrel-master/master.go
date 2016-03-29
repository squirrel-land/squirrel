package main

import (
	"errors"
	"log"
	"net"

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
	log.Printf("%v joined\n", ipAddr)
}

func (master *Master) clientLeave(identity int, err error) {
	master.addrReverse.Remove(master.clients[identity].Addr)
	master.clients[identity] = nil
	master.positionManager.Disable(identity)
	addr, _ := master.addressPool.GetAddress(identity)
	if err == nil {
		log.Printf("link to %v is terminated with no error\n", addr)
	} else {
		log.Printf("link to %v is terminated with error: %v\n", addr, err)
	}
	log.Printf("%v left\n", addr)
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
		buf        *common.ReusableSlice
		ok         bool
		underlying = make([]int, master.addressPool.Capacity()+1)
	)

	for {
		buf, ok = master.clients[myIdentity].Link.ReadFrame()
		if !ok {
			break
		}
		dst := waterutil.MACDestination(buf.Slice())
		if waterutil.IsBroadcast(dst) || waterutil.IsIPv4Multicast(dst) {
			recipients := master.september.SendBroadcast(myIdentity, len(buf.Slice()), underlying)
			for _, id := range recipients {
				if master.clients[id] != nil {
					buf.AddOwner()
					master.clients[id].Link.WriteFrame(buf)
					if *debug {
						log.Printf("broadcast frame of length %d from client %d to be delivered to client %d\n", len(buf.Slice()), myIdentity, id)
					}
				}
			}
			buf.Done()
		} else { // unicast
			dstID, ok := master.addrReverse.Get(waterutil.MACDestination(buf.Slice()))
			if ok {
				if master.september.SendUnicast(myIdentity, dstID, len(buf.Slice())) {
					master.clients[dstID].Link.WriteFrame(buf)
					if *debug {
						log.Printf("unicast frame of length %d from client %d to be delivered to client %d\n", len(buf.Slice()), myIdentity, dstID)
					}
				} else {
					buf.Done()
					if *debug {
						log.Printf("unicast frame of length %d from client %d NOT to be delivered to client %d\n", len(buf.Slice()), myIdentity, dstID)
					}
				}
			} else {
				if *debug {
					log.Printf("unicast frame of length %d from client %d has unknown dst address: %v\n", len(buf.Slice()), myIdentity, waterutil.MACDestination(buf.Slice()))
				}
			}
		}
	}
	master.clientLeave(myIdentity, master.clients[myIdentity].Link.IncomingError())
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
