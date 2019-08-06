package capacity

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var sizeSuffixMulitiplyer = map[string]int{
	"TiB": 1099511627776,
	"TB":  1000000000000,
	"GiB": 1073741824,
	"GB":  1000000000,
	"MiB": 1048576,
	"MB":  1000000,
}

func parseSize(raw string) (int, error) {
	intVal, err := strconv.Atoi(raw)
	if err == nil {
		// specified raw bytes
		return intVal, nil
	}
	for suffix, multiplyer := range sizeSuffixMulitiplyer {
		if strings.HasSuffix(raw, suffix) {
			rawInt := strings.TrimSuffix(raw, suffix)
			intVal, err := strconv.Atoi(rawInt)
			if err != nil {
				return 0, err
			}
			return intVal * multiplyer, nil
		}
	}
	return 0, fmt.Errorf("unable to parse size: %s", raw)
}

func ParseCapacityBytes(raw string) (string, int, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return "", 0, errors.New("must format capacity correctly and include pool")
	}
	pool := parts[0]
	rawCapacity := parts[1]
	sizeBytes, err := parseSize(rawCapacity)
	if err != nil {
		return "", 0, err
	}
	return pool, sizeBytes, nil
}
