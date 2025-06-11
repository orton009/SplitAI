package apiServer

type SplitJson struct {
	Type            string
	TotalAmount     float64
	EqualSplit      bool
	PercentageSplit map[string]float64
	ShareSplit      map[string]int
	UnitSplit       map[string]float64
}
