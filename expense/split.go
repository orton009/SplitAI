package expense

import (
	"encoding/json"
	"fmt"
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
	SplitMap    map[string]Fraction `json:"splitMap"`
	TotalAmount float64             `json:"totalAmount"`
}

func (s *ShareSplit) ComputeTotal() float64 {
	var total float64
	for _, fr := range s.SplitMap {
		total += float64(fr.Numerator/fr.Denominator) * s.TotalAmount
	}

	return total
}

func (s *ShareSplit) GetPayeeSplit() map[string]float64 {
	splitDetail := map[string]float64{}

	for uid, fr := range s.SplitMap {
		splitDetail[uid] = float64(fr.Numerator/fr.Denominator) * s.TotalAmount
	}

	return splitDetail
}

// SplitWrapper handles JSON marshaling/unmarshaling of Split interface
type SplitWrapper struct {
	Split Split           `json:"-"`
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data"`
}

// MarshalJSON custom marshaling for Split interface
func (sw SplitWrapper) MarshalJSON() ([]byte, error) {
	// Marshal the underlying data
	data, err := json.Marshal(sw.Split)
	if err != nil {
		return nil, err
	}

	// Determine the type
	var typeName string
	switch sw.Split.(type) {
	case *EqualSplit:
		typeName = "equal"
	case *UnitSplit:
		typeName = "unit"
	case *PercentageSplit:
		typeName = "percentage"
	case *ShareSplit:
		typeName = "share"
	default:
		return nil, fmt.Errorf("unknown split type: %T", sw.Split)
	}

	// Create wrapper JSON
	return json.Marshal(struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}{
		Type: typeName,
		Data: data,
	})
}

// UnmarshalJSON custom unmarshaling for Split interface
func (sw *SplitWrapper) UnmarshalJSON(data []byte) error {
	// First unmarshal to get the type
	var temp struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Create appropriate concrete type based on type field
	switch temp.Type {
	case "equal":
		var es EqualSplit
		if err := json.Unmarshal(temp.Data, &es); err != nil {
			return err
		}
		sw.Split = &es
	case "unit":
		var us UnitSplit
		if err := json.Unmarshal(temp.Data, &us); err != nil {
			return err
		}
		sw.Split = &us
	case "percentage":
		var ps PercentageSplit
		if err := json.Unmarshal(temp.Data, &ps); err != nil {
			return err
		}
		sw.Split = &ps
	case "share":
		var ss ShareSplit
		if err := json.Unmarshal(temp.Data, &ss); err != nil {
			return err
		}
		sw.Split = &ss
	default:
		return fmt.Errorf("unknown split type: %s", temp.Type)
	}

	sw.Type = temp.Type
	return nil
}
