package main

import (
	"flag"
	"fmt"
)

// Flags
var (
	fServerAddr string
	fIdentity   int
	fTapName    string
)

func init() {
	flag.StringVar(&fServerAddr, "m", "", "master URI")
	flag.IntVar(&fIdentity, "i", 0, "Identity if running as a client")
	flag.StringVar(&fTapName, "t", "", "TAP interface name; leave as blank for default")
}

func main() {
	flag.Parse()
	if fServerAddr == "" || fIdentity == 0 {
		flag.PrintDefaults()
	}

	client, err := NewClient(fTapName)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	err = client.Run(fServerAddr, fIdentity)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}
