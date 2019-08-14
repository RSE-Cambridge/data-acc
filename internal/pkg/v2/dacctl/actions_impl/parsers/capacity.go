package parsers

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// TODO: missing a few?
var sizeSuffixMultiplier = map[string]int{
	"TiB": 1099511627776,
	"TB":  1000000000000,
	"GiB": 1073741824,
	"GB":  1000000000,
	"MiB": 1048576,
	"MB":  1000000,
}

func ParseSize(raw string) (int, error) {
	intVal, err := strconv.Atoi(raw)
	if err == nil {
		// specified raw bytes
		return intVal, nil
	}
	for suffix, multiplier := range sizeSuffixMultiplier {
		if strings.HasSuffix(raw, suffix) {
			rawInt := strings.TrimSpace(strings.TrimSuffix(raw, suffix))
			floatVal, err := strconv.ParseFloat(rawInt, 64)
			if err != nil {
				return 0, err
			}
			floatBytes := floatVal * float64(multiplier)
			return int(math.Ceil(floatBytes)), nil
		}
	}
	return 0, fmt.Errorf("unable to parse size: %s", raw)
}

func ParseCapacityBytes(raw string) (string, int, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return "", 0, errors.New("must format capacity correctly and include pool")
	}
	pool := strings.TrimSpace(parts[0])
	rawCapacity := strings.TrimSpace(parts[1])
	sizeBytes, err := ParseSize(rawCapacity)
	if err != nil {
		return "", 0, err
	}
	return pool, sizeBytes, nil
}
