// Package uti provides general, low priority methods and functions for different purposes
package uti

import (
	"math"
	"strconv"
	"strings"
)

// Round serves the rol of the missing math.Round function
func Round(val float64, roundOn float64, places int) (newVal float64) {

	var round float64

	pow := math.Pow(10, float64(places))
	digit := pow * val

	_, div := math.Modf(digit)
	_div := math.Copysign(div, val)
	_roundOn := math.Copysign(roundOn, val)

	if _div >= _roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}

	newVal = round / pow

	return
}

// ToFixed truncates float numbers to a specific position
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(toFixedRound(num*output)) / output
}

func toFixedRound(num float64) int {
	return int(num + math.Copysign(0.05, num))
}

// ParseFloat converts scientific notation values from string format to float64
func ParseFloat(str string) (float64, error) {

	val, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return val, nil
	}

	//Some number may be seperated by comma, for example, 23,120,123, so remove the comma firstly
	str = strings.Replace(str, ",", "", -1)

	//Some number is specifed in scientific notation
	pos := strings.IndexAny(str, "eE")
	if pos < 0 {
		return strconv.ParseFloat(str, 64)
	}

	var baseVal float64
	var expVal int64

	baseStr := str[0:pos]
	baseVal, err = strconv.ParseFloat(baseStr, 64)
	if err != nil {
		return 0, err
	}

	expStr := str[(pos + 1):]
	expVal, err = strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return baseVal * math.Pow10(int(expVal)), nil
}
