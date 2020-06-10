// Package hamradio provides the defintion of the most basic data types that are used in amateur radio.
package hamradio

import (
	"fmt"
	"math"
)

// Frequency represents a frequency in Hz.
type Frequency float64

func (f Frequency) String() string {
	return fmt.Sprintf("%.2fHz", f)
}

// FrequencyRange represents a range of frequencies.
type FrequencyRange struct {
	From, To Frequency
}

func (r FrequencyRange) String() string {
	return fmt.Sprintf("[%v, %v]", r.From, r.To)
}

// Center frequency of this range.
func (r FrequencyRange) Center() Frequency {
	return r.From + (r.To-r.From)/2
}

// Width of this frequency range in Hz.
func (r FrequencyRange) Width() Frequency {
	return r.To - r.From
}

// Contains indicates if the given frequency is within this frequency range.
func (r FrequencyRange) Contains(f Frequency) bool {
	return f >= r.From && f <= r.To
}

// Shift this frequency range by the given Δ.
func (r *FrequencyRange) Shift(Δ Frequency) {
	r.From += Δ
	r.To += Δ
}

// Expanded returns a new frequency range expanded by the given Δ.
func (r FrequencyRange) Expanded(Δ Frequency) FrequencyRange {
	return FrequencyRange{From: r.From - Δ, To: r.To + Δ}
}

// DB represents decibel (dB).
type DB float64

func (l DB) String() string {
	return fmt.Sprintf("%.2fdB", l)
}

// ToSUnit converts this value in dB into the corresponding S-unit (S0 - S9).
func (l DB) ToSUnit() (s int, unit SUnit, add DB) {
	for i := len(SUnits) - 1; i >= 0; i-- {
		if l >= DB(SUnits[i]) {
			s = i
			unit = SUnits[i]
			add = l - DB(unit)
			return s, unit, add
		}
	}
	return 0, S0, l - DB(S0)
}

// SUnit represents the upper bound of a S-unit in dBm.
type SUnit DB

const (
	S0 SUnit = -127
	S1 SUnit = -121
	S2 SUnit = -115
	S3 SUnit = -109
	S4 SUnit = -103
	S5 SUnit = -97
	S6 SUnit = -91
	S7 SUnit = -85
	S8 SUnit = -79
	S9 SUnit = -73
)

// SUnits contains all S-units (S0 - S9).
var SUnits = []SUnit{S0, S1, S2, S3, S4, S5, S6, S7, S8, S9}

func (u SUnit) String() string {
	s, _, add := DB(u).ToSUnit()
	if s == 9 {
		return fmt.Sprintf("S%d+%.0fdB", s, add)
	} else if s > 0 {
		return fmt.Sprintf("S%d", s)
	} else {
		return fmt.Sprintf("S%d%.0fdB", s, add)
	}
}

// DBRange represents a range of dB.
type DBRange struct {
	From, To DB
}

func (r DBRange) String() string {
	return fmt.Sprintf("[%v,%v]", r.From, r.To)
}

// Normalized returns a normalized version of this dB range, where From <= To.
func (r DBRange) Normalized() DBRange {
	if r.From > r.To {
		return DBRange{
			From: r.To,
			To:   r.From,
		}
	}
	return r
}

// Width of this dB range.
func (r DBRange) Width() DB {
	return DB(math.Abs(float64(r.To - r.From)))
}

// Contains indicates if the given value in dB is within this dB range.
func (r DBRange) Contains(value DB) bool {
	return value >= r.From && value <= r.To
}
