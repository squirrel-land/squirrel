package main

import (
	"flag"
	"fmt"
	"net"
)

// Flags
var (
	fConfig string
)

func init() {
	flag.StringVar(&fConfig, "c", "", "path to configuration file.")
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
	master := NewMaster(network, mobilityManager, september)
	return master.Run(config.ListenAddress)
}

func main() {
	flag.Parse()
	if fConfig == "" {
		flag.PrintDefaults()
	} else {
		err := runMaster()
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}
}
