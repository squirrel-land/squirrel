package simpleModels

import (
	"../modelDep"
)

type staticUniformPositions struct {
	nodes          []*modelDep.Position
	spacing        float64
	next           func(*modelDep.Position, *modelDep.Position, float64)
	latestPosition *modelDep.Position
}

func NewStaticUniformPositions() modelDep.MobilityManager {
	return &staticUniformPositions{}
}

func (mobilityManager *staticUniformPositions) ParametersHelp() string {
	return ""
}

func (mobilityManager *staticUniformPositions) Configure(config map[string]interface{}) error {
	spacing, ok := config["Spacing"].(float64)
	if ok != true {
		return ParametersNotValid
	}
	shape, ok := config["Shape"].(string)
	if ok != true {
		return ParametersNotValid
	}
	switch shape {
	case "Linear":
		mobilityManager.next = staticNextPointLinear
	default:
		return ParametersNotValid
	}
	mobilityManager.spacing = spacing
	return nil
}

func (mobilityManager *staticUniformPositions) GenerateNewNode() (newPosition *modelDep.Position) {
	newPosition = &modelDep.Position{0, 0, 0}
	mobilityManager.next(mobilityManager.latestPosition, newPosition, mobilityManager.spacing)
	mobilityManager.latestPosition = newPosition
	return
}

func (mobilityManager *staticUniformPositions) Initialize(nodes []*modelDep.Position) {
	mobilityManager.nodes = nodes
}

func staticNextPointLinear(prev *modelDep.Position, next *modelDep.Position, spacing float64) {
	if prev == nil {
		next.X = 0
		next.Y = 0
		next.Height = 0
	} else {
		next.X = prev.X + spacing
		next.Y = prev.Y
		next.Height = prev.Height
	}
}
