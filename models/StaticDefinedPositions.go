package models

import (
	"github.com/songgao/squirrel/models/common"
)

type staticDefinedPositions struct {
	positions [][3]float64
}

func newStaticDefinedPositions() common.MobilityManager {
	return &staticDefinedPositions{}
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
	mobilityManager.positions = pos
	return nil
}

func (mobilityManager *staticDefinedPositions) Initialize(positionManager *common.PositionManager) {
	ch := make(chan []int)
	positionManager.RegisterEnabledChanged(ch)
	go func() {
		for {
			enabled := <-ch
			for i, index := range enabled {
				if i < len(mobilityManager.positions) {
					p := mobilityManager.positions[i]
					positionManager.Set(index, p[0], p[1], p[2])
				}
			}
		}
	}()
}
