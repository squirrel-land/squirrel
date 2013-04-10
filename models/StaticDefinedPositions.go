package models

import (
	"github.com/songgao/squirrel/models/common"
)

type staticDefinedPositions struct {
	nodes     []*common.Position
	positions chan *common.Position
}

func newStaticDefinedPositions() common.MobilityManager {
	return &staticDefinedPositions{positions: make(chan *common.Position)}
}

func (mobilityManager *staticDefinedPositions) ParametersHelp() string {
	return ``
}

func (mobilityManager *staticDefinedPositions) Configure(config map[string]interface{}) error {
	positions, ok := config["Positions"].([]interface{})
	if ok != true {
		return ParametersNotValid
	}
	pos := make([][3]float64, len(positions))
	for i := range positions {
		position, ok := positions[i].([]interface{})
		if ok != true {
			return ParametersNotValid
		}
		for j := 0; j < 3; j++ {
			num, ok := position[j].(float64)
			if ok != true {
				return ParametersNotValid
			}
			pos[i][j] = num
		}
	}
	go func() {
		for i := range positions {
			mobilityManager.positions <- common.NewPositionFromArray(pos[i])
		}
		mobilityManager.positions <- nil
	}()
	return nil
}

func (mobilityManager *staticDefinedPositions) GenerateNewNode() (newPosition *common.Position) {
	pos := <-mobilityManager.positions
	if pos == nil {
		mobilityManager.positions <- nil
	}
	return pos
}

func (mobilityManager *staticDefinedPositions) Initialize(nodes []*common.Position) {
	mobilityManager.nodes = nodes
}
