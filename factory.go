package main

import (
	"errors"
	"github.com/songgao/squirrel/models"
	"github.com/songgao/squirrel/models/common"
)

var (
	notRegistered = errors.New("MobilityManager or September is not registered.")
)

func newMobilityManager(name string) (mobilityManager common.MobilityManager, err error) {
	constructor := models.MobilityManagers[name]
	if constructor == nil {
		return nil, notRegistered
	}
	mobilityManager = constructor()
	return
}

func newSeptember(name string) (september common.September, err error) {
	constructor := models.Septembers[name]
	if constructor == nil {
		return nil, notRegistered
	}
	september = constructor()
	return
}
