package models

import (
	"github.com/songgao/squirrel/models/common"
	"math"
)

func distance(p1 *common.Position, p2 *common.Position) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2) + math.Pow(p1.Height-p2.Height, 2))
}
