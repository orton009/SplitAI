package expense

type Split interface {
	computeTotal() float64
	getPayeeSplit() map[int]float64
}

type EqualSplit struct {
	Payee       []int
	TotalAmount float64
}

func (e *EqualSplit) computeTotal() float64 {
	return e.TotalAmount
}

func (e *EqualSplit) getPayeeSplit() map[int]float64 {
	amountSplit := make(map[int]float64)
	amount := e.TotalAmount / float64(len(e.Payee))
	for i := range e.Payee {
		amountSplit[e.Payee[i]] = amount
	}
	return amountSplit
}

type UnitSplit struct {
	PayeeAmountSplit map[int]float64
}

func (u *UnitSplit) computeTotal() float64 {
	var totalAmount float64
	for _, amount := range u.PayeeAmountSplit {
		totalAmount += amount
	}

	return totalAmount
}

func (u *UnitSplit) getPayeeSplit() map[int]float64 {
	return u.PayeeAmountSplit
}

type PercentageSplit struct {
	PercentageSplitMap map[int]float64
	TotalAmount        float64
}

func (p *PercentageSplit) getPayeeSplit() map[int]float64 {
	splitMap := map[int]float64{}
	for userId, percent := range p.PercentageSplitMap {
		splitMap[userId] = roundFloat((percent/100)*p.TotalAmount, 2)
	}
	return splitMap
}

func (p *PercentageSplit) computeTotal() float64 {
	var totalPercent float64
	for _, percent := range p.PercentageSplitMap {
		totalPercent += percent
	}

	if totalPercent == 100 {
		return p.TotalAmount
	} else {
		return roundFloat((totalPercent/100)*p.TotalAmount, 2)
	}
}

type Fraction struct {
	Numerator   int
	Denominator int
}

type ShareSplit struct {
	SplitMap    map[int]Fraction
	TotalAmount float64
}

func (s *ShareSplit) computeTotal() float64 {
	var total float64
	for _, fr := range s.SplitMap {
		total += float64(fr.Numerator/fr.Denominator) * s.TotalAmount
	}

	return total
}

func (s *ShareSplit) getPayeeSplit() map[int]float64 {
	splitDetail := map[int]float64{}

	for uid, fr := range s.SplitMap {
		splitDetail[uid] = float64(fr.Numerator/fr.Denominator) * s.TotalAmount
	}

	return splitDetail
}
