package main

import (
	"errors"

	"github.com/squirrel-land/models"
	"github.com/squirrel-land/squirrel"
)

var (
	notRegistered = errors.New("MobilityManager or September is not registered.")
)

func newMobilityManager(name string) (mobilityManager squirrel.MobilityManager, err error) {
	constructor := models.MobilityManagers[name]
	if constructor == nil {
		return nil, notRegistered
	}
	mobilityManager = constructor()
	return
}

func newSeptember(name string) (september squirrel.September, err error) {
	constructor := models.Septembers[name]
	if constructor == nil {
		return nil, notRegistered
	}
	september = constructor()
	return
}
