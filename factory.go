package main

import (
	"./modelDep"
	"errors"
)

// To avoid imports in constructors.go
type (
	typeMobilityManagerConstructor func() modelDep.MobilityManager
	typeSeptemberConstructor       func() modelDep.September
)

var (
	notRegistered = errors.New("MobilityManager or September is not registered.")
)

func newMobilityManager(name string) (mobilityManager modelDep.MobilityManager, err error) {
	constructor := mobilityManagers[name]
	if constructor == nil {
		return nil, notRegistered
	}
	mobilityManager = constructor()
	return
}

func newSeptember(name string) (september modelDep.September, err error) {
	constructor := septembers[name]
	if constructor == nil {
		return nil, notRegistered
	}
	september = constructor()
	return
}
