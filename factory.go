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
	NotRegistered = errors.New("MobilityManager or September is not registered.")
)

func NewMobilityManager(name string) (mobilityManager master.MobilityManager, err error) {
	constructor := mobilityManagers[name]
	if constructor == nil {
		return nil, NotRegistered
	}
	mobilityManager = constructor()
	return
}

func NewSeptember(name string) (september master.September, err error) {
	constructor := septembers[name]
	if constructor == nil {
		return nil, NotRegistered
	}
	september = constructor()
	return
}
