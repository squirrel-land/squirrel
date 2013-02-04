package main

import (
	"github.com/songgao/squirrel/client"
	"github.com/songgao/squirrel/master"
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
	mobilityManager, err := newMobilityManager(config.MobilityManager)
	if err != nil {
		return
	}
	err = mobilityManager.Configure(config.MobilityManagerParameters)
	if err != nil {
		fmt.Println("Creating MobilityManager failed. Following message might help:\n")
		fmt.Println(mobilityManager.ParametersHelp())
		return
	}
	september, err := newSeptember(config.September)
	if err != nil {
		return
	}
	err = september.Configure(config.SeptemberParameters)
	if err != nil {
		fmt.Println("Creating September failed. Following message might help:\n")
		fmt.Println(september.ParametersHelp())
		return
	}
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
