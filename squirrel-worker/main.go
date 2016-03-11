package main

import (
	"fmt"
	"log"
	"os"

	"github.com/coreos/go-etcd/etcd"
	"github.com/squirrel-land/squirrel/common"
)

type config struct {
	masterURI string
	tapName   string
}

func getConfig() (conf config, err error) {
	endpoint := os.Getenv("SQUIRREL_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:4001"
	}
	client := etcd.NewClient([]string{endpoint})

	conf.masterURI, err = common.GetEtcdValue(client, "/squirrel/master_uri")
	if err != nil {
		return
	}

	conf.tapName, err = common.GetEtcdValue(client, "/squirrel/worker_tap_name")
	if err != nil {
		if common.IsEtcdNotFoundError(err) {
			// syscalls in `water` uses default TAP interface name if empty
			err = nil
		} else {
			return
		}
	}

	return
}

func printHelp() {
	fmt.Println()
	fmt.Printf("Usage: %s\n", os.Args[0])
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("    SQUIRREL_ENDPOINT  : etcd endpoint UIR. [Optional]")
	fmt.Println("                             Default: http://127.0.0.1:4001")
	fmt.Println()
	fmt.Println("Etcd Configuration Entries:")
	fmt.Println("    /squirrel/master_uri      : URI of the squirrel-master. [Required]")
	fmt.Println("    /squirrel/worker_tap_name : Name of the TAP interface.  [Optional]")
}

func main() {
	conf, err := getConfig()

	if err != nil {
		log.Println(err)
		printHelp()
		os.Exit(1)
	}

	client, err := NewClient(conf.tapName)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = client.Run(conf.masterURI)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	return
}
