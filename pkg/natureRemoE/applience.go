package natureRemoE

import (
	"fmt"
	"math"
	"strconv"
)

const (
	EpcMeasuredInstantaneous                    = 231
	EpcNormalDirectionCumulativeElectricEnergy  = 224
	EpcReverseDirectionCumulativeElectricEnergy = 227
	EpcCumulativeElectricEnergyUnit             = 225
	EpcCoefficient                              = 211
	EpcCumulativeElectricEnergyEffectiveDigits  = 215
)

type echonetliteProperty struct {
	Name      string `json:"name"`
	Epc       int    `json:"epc"`
	Val       string `json:"val"`
	UpdatedAt string `json:"updated_at"`
}

type Applience struct {
	Id         string `json:"id"`
	SmartMeter struct {
		EchonetliteProperties []echonetliteProperty `json:"echonetlite_properties"`
	} `json:"smart_meter"`
}
type Energy struct {
	Timestamp         string
	Instantaneous     int
	NormalCumulative  float64
	ReverseCumulative float64
	OverflowDiff      float64
}

func ParseEnergy(a Applience) (Energy, error) {
	var e Energy
	var err error

	var normalCumulativeBase int
	var reverseCumulativeBase int
	var coefficient int
	var unit float64
	var effectiveDigits int

	for _, p := range a.SmartMeter.EchonetliteProperties {
		switch p.Epc {
		case EpcMeasuredInstantaneous:
			if e.Instantaneous, err = strconv.Atoi(p.Val); err != nil {
				return e, fmt.Errorf("Error: failed to parse Measured Instantaneous (EPC: %d)", p.Epc)
			}
			// assume "updated_at" in instantaneous to be timestamp
			e.Timestamp = p.UpdatedAt
		case EpcNormalDirectionCumulativeElectricEnergy:
			if normalCumulativeBase, err = strconv.Atoi(p.Val); err != nil {
				return e, fmt.Errorf("Error: failed to parse Normal Direction Cumulative Electric Energy (EPC: %d)", p.Epc)
			}
		case EpcReverseDirectionCumulativeElectricEnergy:
			if reverseCumulativeBase, err = strconv.Atoi(p.Val); err != nil {
				return e, fmt.Errorf("Error: failed to parse Reverse Direction Cumulative Electric Energy (EPC: %d)", p.Epc)
			}
		case EpcCumulativeElectricEnergyUnit:
			switch p.Val {
			case "0": // 0x00
				unit = 1.0
			case "1": // 0x01
				unit = 0.1
			case "2": // 0x02
				unit = 0.01
			case "3": // 0x03
				unit = 0.001
			case "4": // 0x03
				unit = 0.0001
			case "10": // 0x0A
				unit = 10.0
			case "11": // 0x0B
				unit = 100.0
			case "12": // 0x0C
				unit = 1000.0
			case "13": // 0x0D
				unit = 10000.0
			default:
				return e, fmt.Errorf("Error: failed to parse Energy Unit (EPC: %d)", p.Epc)
			}
		case EpcCoefficient:
			if coefficient, err = strconv.Atoi(p.Val); err != nil {
				return e, fmt.Errorf("Error: failed to parse Coefficient (EPC: %d)", p.Epc)
			}
		case EpcCumulativeElectricEnergyEffectiveDigits:
			if effectiveDigits, err = strconv.Atoi(p.Val); err != nil {
				return e, fmt.Errorf("Error: failed to parse Effective Digits (EPC: %d)", p.Epc)
			}
		}
	}

	e.NormalCumulative = float64(normalCumulativeBase*coefficient) * unit
	e.ReverseCumulative = float64(reverseCumulativeBase*coefficient) * unit
	e.OverflowDiff = math.Pow10(effectiveDigits) * float64(coefficient) * unit

	return e, nil
}
