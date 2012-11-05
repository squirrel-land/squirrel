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
	fMaster bool
	fClient bool
	fConfig string
)

func init() {
	flag.BoolVar(&fMaster, "m", false, "Run as master.")
	flag.BoolVar(&fClient, "c", false, "Run as client.")
	flag.StringVar(&fConfig, "f", "", "Configuration file.")
}

func runMaster() (err error) {
	config, err := parseMasterConfig(fConfig)
	if err != nil {
		return
	}
	_, network, err := net.ParseCIDR(config.Network)
	if err != nil {
		return
	}
	mobilityManager, _ := NewMobilityManager(config.MobilityManager, config.MobilityManagerParameters)
	september, _ := NewSeptember(config.September, config.SeptemberParameters)
	master := master.NewMaster(network, mobilityManager, september)
	return master.Run(config.ListenAddress)
}

func runClient() (err error) {
	config, err := parseClientConfig(fConfig)
	if err != nil {
		return
	}
	client, err := client.NewClient(config.TunInterfaceName)
	if err != nil {
		return
	}
	return client.Run(config.ServerAddress, config.Identity)
}

func main() {
	flag.Parse()
	switch {
	case fMaster:
		err := runMaster()
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	case fClient:
		err := runClient()
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	default:
		flag.PrintDefaults()
	}
}
