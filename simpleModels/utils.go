package simpleModels

import (
	"../modelDep"
	"math"
)

func distance(p1 *modelDep.Position, p2 *modelDep.Position) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2) + math.Pow(p1.Height-p2.Height, 2))
}
