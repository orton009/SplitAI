package expense

import (
	"encoding/json"
	"fmt"

	"github.com/samber/lo"
)

type Split interface {
	ComputeTotal() float64
	GetPayeeSplit() map[string]float64
}

type EqualSplit struct {
	Payee       []string `json:"payee"`
	TotalAmount float64  `json:"totalAmount"`
}

func (e *EqualSplit) ComputeTotal() float64 {
	return e.TotalAmount
}

func (e *EqualSplit) GetPayeeSplit() map[string]float64 {
	amountSplit := make(map[string]float64)
	amount := e.TotalAmount / float64(len(e.Payee))
	for i := range e.Payee {
		amountSplit[e.Payee[i]] = amount
	}
	return amountSplit
}

type UnitSplit struct {
	PayeeAmountSplit map[string]float64 `json:"payeeAmountSplit"`
}

func (u *UnitSplit) ComputeTotal() float64 {
	var totalAmount float64
	for _, amount := range u.PayeeAmountSplit {
		totalAmount += amount
	}

	return totalAmount
}

func (u *UnitSplit) GetPayeeSplit() map[string]float64 {
	return u.PayeeAmountSplit
}

type PercentageSplit struct {
	PercentageSplitMap map[string]float64 `json:"percentageSplitMap"`
	TotalAmount        float64            `json:"totalAmount"`
}

func (p *PercentageSplit) GetPayeeSplit() map[string]float64 {
	splitMap := map[string]float64{}
	for userId, percent := range p.PercentageSplitMap {
		splitMap[userId] = roundFloat((percent/100)*p.TotalAmount, 2)
	}
	return splitMap
}

func (p *PercentageSplit) ComputeTotal() float64 {
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
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
}

type ShareSplit struct {
	SplitMap    map[string]int `json:"splitMap"`
	TotalAmount float64        `json:"totalAmount"`
}

func (s *ShareSplit) ComputeTotal() float64 {
	var total float64
	var totalShares = lo.Sum(lo.Values(s.SplitMap))
	for _, share := range s.SplitMap {
		total += float64(share) / float64(totalShares) * s.TotalAmount
	}
	return total
}

func (s *ShareSplit) GetPayeeSplit() map[string]float64 {
	splitDetail := map[string]float64{}

	var totalShares = lo.Sum(lo.Values(s.SplitMap))
	for uid, fr := range s.SplitMap {
		splitDetail[uid] = float64(fr/totalShares) * s.TotalAmount
	}

	return splitDetail
}

// SplitWrapper handles JSON marshaling/unmarshaling of Split interface
type SplitWrapper struct {
	Split Split  `json:"-"`
	Type  string `json:"type"`
}

// MarshalJSON custom marshaling for Split interface
func (sw SplitWrapper) MarshalJSON() ([]byte, error) {

	var sj SplitJson

	// Determine the type
	switch sw.Split.(type) {
	case *EqualSplit:
		sj.Type = "equal"
		sj.EqualSplit = sw.Split.(*EqualSplit).Payee
		sj.TotalAmount = sw.Split.(*EqualSplit).TotalAmount
	case *UnitSplit:
		sj.Type = "unit"
		sj.UnitSplit = sw.Split.(*UnitSplit).PayeeAmountSplit
	case *PercentageSplit:
		sj.Type = "percentage"
		sj.PercentageSplit = sw.Split.(*PercentageSplit).PercentageSplitMap
		sj.TotalAmount = sw.Split.(*PercentageSplit).TotalAmount
	case *ShareSplit:
		sj.Type = "share"
		sj.ShareSplit = sw.Split.(*ShareSplit).SplitMap
		sj.TotalAmount = sw.Split.(*ShareSplit).TotalAmount
	default:
		return nil, fmt.Errorf("unknown split type: %T", sw.Split)
	}

	return json.Marshal(sj)

}

// UnmarshalJSON custom unmarshaling for Split interface
func (sw *SplitWrapper) UnmarshalJSON(data []byte) error {
	// First unmarshal to get the type
	var temp SplitJson
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Create appropriate concrete type based on type field
	switch temp.Type {
	case "equal":
		var es EqualSplit
		es.Payee = temp.EqualSplit
		es.TotalAmount = temp.TotalAmount
		sw.Split = &es
	case "unit":
		var us UnitSplit
		us.PayeeAmountSplit = temp.UnitSplit
		sw.Split = &us
	case "percentage":
		var ps PercentageSplit
		ps.TotalAmount = temp.TotalAmount
		ps.PercentageSplitMap = temp.PercentageSplit
		sw.Split = &ps
	case "share":
		var ss ShareSplit
		ss.TotalAmount = temp.TotalAmount
		ss.SplitMap = temp.ShareSplit
		sw.Split = &ss
	default:
		return fmt.Errorf("unknown split type: %s", temp.Type)
	}

	sw.Type = temp.Type
	return nil
}
