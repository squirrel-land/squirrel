package main

import (
	"./master"
	"errors"
)

// To avoid imports in constructors.go
type (
	typeMobilityManagerConstructor func() master.MobilityManager
	typeSeptemberConstructor       func() master.September
)

var (
	notRegistered = errors.New("MobilityManager or September is not registered.")
)

func newMobilityManager(name string) (mobilityManager master.MobilityManager, err error) {
	constructor := mobilityManagers[name]
	if constructor == nil {
		return nil, notRegistered
	}
	mobilityManager = constructor()
	return
}

func newSeptember(name string) (september master.September, err error) {
	constructor := septembers[name]
	if constructor == nil {
		return nil, notRegistered
	}
	september = constructor()
	return
}
