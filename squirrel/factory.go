package main

import (
	"./master"
	"errors"
)

// To avoid imports in constructors.go
type (
	typeMobilityManagerConstructor func(map[string]string) (master.MobilityManager, error)
	typeSeptemberConstructor       func(map[string]string) (master.September, error)
)

var (
	NotRegistered = errors.New("MobilityManager or September is not registered.")
)

func NewMobilityManager(name string, config map[string]string) (mobilityManager master.MobilityManager, err error) {
	constructor := mobilityManagers[name]
	if constructor == nil {
		return nil, NotRegistered
	}
	mobilityManager, err = constructor(config)
	return
}

func NewSeptember(name string, config map[string]string) (september master.September, err error) {
	constructor := septembers[name]
	if constructor == nil {
		return nil, NotRegistered
	}
	september, err = constructor(config)
	return
}
