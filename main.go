package main

import (
	"flag"
	"fmt"
	"github.com/songgao/squirrel/client"
	"github.com/songgao/squirrel/master"
	"net"
)

// Flags
var (
	fMaster     string
	fServerAddr string
	fIdentity   int
)

func init() {
	flag.StringVar(&fMaster, "m", "", "Master mode; Configuration file.")
	flag.StringVar(&fServerAddr, "c", "", "Run as client; server URI")
	flag.IntVar(&fIdentity, "i", 0, "Identity if running as a clieng")
}

func runMaster() (err error) {
	config, err := parseMasterConfig(fMaster)
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
	client, err := client.NewClient("")
	if err != nil {
		return
	}
	return client.Run(fServerAddr, fIdentity)
}

func main() {
	flag.Parse()
	switch {
	case fMaster != "":
		err := runMaster()
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	case fServerAddr != "" && fIdentity != 0:
		err := runClient()
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	default:
		flag.PrintDefaults()
	}
}
