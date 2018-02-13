package meta

import (
	"fmt"
	"strconv"
	"strings"
)

var levels = map[string]float64{
		"debug":   float64(0),
		"info":    float64(1),
		"warn":    float64(2),
		"warning": float64(2),
		"error":   float64(3),
		"fatal":   float64(4),
	}

func getFloatValue(field, str string) (float64, error) {
	if strings.ToLower(field) == "level" {
		level, ok := levels[strings.ToLower(str)]
		if !ok {
			return level, fmt.Errorf("Level %s not defined", str)
		}
		return level, nil
	}
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return v, fmt.Errorf("Can not parse float of value %s", str)
	}
	return v, nil
}
