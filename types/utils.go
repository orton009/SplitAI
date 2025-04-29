package expense

import "math"

func roundFloat(n float64, precision int) float64 {
	return math.Floor(n*math.Pow10(precision)) / math.Pow10(precision)
}
