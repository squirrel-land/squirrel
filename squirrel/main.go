package main

import (
	"./client"
	"./master"
	"flag"
	"fmt"
	"net"
)

// Flags
var (
	fMaster   bool
	fClient   bool
	fNetwork  string
	fAddr     string
	fIdentity int
	fTunName  string
)

func init() {
	flag.BoolVar(&fMaster, "m", false, "Run as master.")
	flag.BoolVar(&fClient, "c", false, "Run as client.")
	flag.StringVar(&fNetwork, "net", "10.0.0.0/24", "The virtual IP network used in address pool to assign IP address to clients. It's used only when running as master.")
	flag.StringVar(&fAddr, "addr", "", "The corresponding address. If running as master, it's the address to listen on; if running as client, it's the address of the master to connect to. Format: HOST:PORT")
	flag.IntVar(&fIdentity, "id", -1, "The client identity used to differenciate different clients. It's used only when running as client.")
	flag.StringVar(&fTunName, "tun", "", "The TUN interface name used for virtual network. It's used only when running as client.")
}

func isMaster() bool {
	if fClient || !fMaster {
		return false
	}
	if fAddr == "" {
		return false
	}
	return true
}

func isClient() bool {
	if fMaster || !fClient {
		return false
	}
	if fAddr == "" {
		return false
	}
	if fIdentity <= 0 {
		return false
	}
	if fTunName == "" {
		return false
	}
	return true
}

func runMaster() (err error) {
	_, network, err := net.ParseCIDR(fNetwork)
	if err != nil {
		return
	}
	mobilityManager, _ := NewMobilityManager("SimpleMobilityManager", nil)
	september, _ := NewSeptember("September1st", nil)
	master := master.NewMaster(network, mobilityManager, september)
	return master.Run(fAddr)
}

func runClient() (err error) {
	client, err := client.NewClient(fTunName)
	if err != nil {
		return
	}
	return client.Run(fAddr, fIdentity)
}

func main() {
	flag.Parse()
	switch {
	case isMaster():
		err := runMaster()
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	case isClient():
		err := runClient()
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	default:
		flag.PrintDefaults()
	}
}
