package client

import (
	"github.com/songgao/squirrel/common"
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	cINITIAL_ROUTING_TABLE_CACHE_SIZE = 64
	cUPDATE_INTERVAL                  = 0.050
)

var _ZEROS_IP net.IP = net.IPv4(0, 0, 0, 0)

type route struct {
	Network net.IPNet
	Gateway net.IP
}

type routeArray []*route

type routes struct {
	routes      routeArray
	ifName      string
	updatedTime time.Time
}

// Create a new routing table monitor that monitors routes on the interface ifName
func newRoutes(ifName string) (ret *routes) {
	ret = &routes{ifName: ifName}
	ret.update()
	return
}

func (r *routes) initroutes() {
	r.routes = make([]*route, 0, cINITIAL_ROUTING_TABLE_CACHE_SIZE)
}

func (r routeArray) Len() int {
	return len(r)
}

func (r routeArray) Less(i, j int) bool {
	len_i, _ := r[i].Network.Mask.Size()
	len_j, _ := r[j].Network.Mask.Size()
	return len_i > len_j //Descending order
}

func (r routeArray) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func decodeIPHex(ipHexStr string) (ip net.IP, err error) {
	bytes, err := hex.DecodeString(ipHexStr)
	if err != nil {
		return
	}
	ip = common.IPFromBigEndian(bytes)
	return
}

func decodeMaskHex(ipHexStr string) (mask net.IPMask, err error) {
	bytes, err := hex.DecodeString(ipHexStr)
	if err != nil {
		return
	}
	mask = common.IPMaskFromBigEndian(bytes)
	return
}

func (r *routes) update() (err error) {
	file, err := os.OpenFile("/proc/net/route", os.O_RDONLY, 0)
	if err != nil {
		fmt.Println("Open /proc/net/route error")
		return
	}
	r.initroutes()
	f := bufio.NewReader(file)
	line, err := f.ReadString('\n')
	line, err = f.ReadString('\n')
	for err == nil && line != "" {
		entries := strings.Split(line, "\t")
		if r.ifName == entries[0] {
			dst, err := decodeIPHex(entries[1])
			next, err := decodeIPHex(entries[2])
			mask, err := decodeMaskHex(entries[7])
			if err != nil {
				return err
			}
			newRoute := route{Network: net.IPNet{dst, mask}, Gateway: next}
			r.routes = append(r.routes, &newRoute)
		}
		line, err = f.ReadString('\n')
	}
	sort.Sort(r.routes)
	if err == io.EOF {
		r.updatedTime = time.Now()
		return nil
	}
	return
}

func (r *routes) Print() {
	fmt.Println("---- routes ----")
	for i := range r.routes {
		fmt.Printf("%v\n", r.routes[i])
	}
	fmt.Println("-- End routes --")
}

// Find the next-hop (gateway) of the given IP address.
func (r *routes) Route(dst net.IP) net.IP {
	if time.Since(r.updatedTime).Seconds() > cUPDATE_INTERVAL {
		r.update()
	}
	for i := range r.routes {
		if r.routes[i].Network.Contains(dst) {
			if r.routes[i].Gateway.Equal(_ZEROS_IP) {
				return dst
			} else {
				return r.routes[i].Gateway
			}
		}
	}
	return dst
}
