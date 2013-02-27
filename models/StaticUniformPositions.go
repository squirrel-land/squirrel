package models

import (
	"github.com/songgao/squirrel/models/common"
)

type staticUniformPositions struct {
	nodes          []*common.Position
	spacing        float64
	next           func(*common.Position, *common.Position, float64)
	latestPosition *common.Position
}

func NewStaticUniformPositions() common.MobilityManager {
	return &staticUniformPositions{}
}

func (mobilityManager *staticUniformPositions) ParametersHelp() string {
	return `StaticUniformPositions is a mobility manager in which nodes are not mobile.
Nodes are positioned uniformly on a grid map.

  "Spacing": float64, required;
             Space between nodes.
  "Shape":   string, required;
             The shape which positions of nodes should follow; can be one of
             ["Linear"].
    `
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

func (mobilityManager *staticUniformPositions) GenerateNewNode() (newPosition *common.Position) {
	newPosition = common.NewPosition()
	mobilityManager.next(mobilityManager.latestPosition, newPosition, mobilityManager.spacing)
	mobilityManager.latestPosition = newPosition
	return
}

func (mobilityManager *staticUniformPositions) Initialize(nodes []*common.Position) {
	mobilityManager.nodes = nodes
}

func staticNextPointLinear(prev *common.Position, next *common.Position, spacing float64) {
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
