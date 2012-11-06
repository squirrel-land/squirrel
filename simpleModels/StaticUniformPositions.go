package simpleModels

import (
	"../master"
	"errors"
)

var (
	ParametersNotValid = errors.New("Parameter(s) are not valid.")
)

type StaticUniformPositions struct {
	nodes          []*master.Position
	spacing        int
	next           func(*master.Position, *master.Position, int)
	latestPosition *master.Position
}

func NewStaticUniformPositions() master.MobilityManager {
	return &StaticUniformPositions{}
}

func (mobilityManager *StaticUniformPositions) ParametersHelp() string {
	return ""
}

func (mobilityManager *StaticUniformPositions) Configure(config map[string]interface{}) error {
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
	mobilityManager.spacing = int(spacing)
	return nil
}

func (mobilityManager *StaticUniformPositions) GenerateNewNode() (newPosition *master.Position) {
	newPosition = &master.Position{0, 0, 0}
	mobilityManager.next(mobilityManager.latestPosition, newPosition, mobilityManager.spacing)
	mobilityManager.latestPosition = newPosition
	return
}

func (mobilityManager *StaticUniformPositions) Initialize(nodes []*master.Position) {
	mobilityManager.nodes = nodes
}

func staticNextPointLinear(prev *master.Position, next *master.Position, spacing int) {
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
