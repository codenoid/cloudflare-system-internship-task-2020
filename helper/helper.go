package helper

import (
	"regexp"
	"strings"
)

var domainRX = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-z]{1,63}| xn--[a-z0-9]{1,59})$`)

// IsDomainName check whenever a give string can be parsed
// and the Host are exist
func IsDomainName(str string) bool {
	return domainRX.MatchString(str)
}

// IsIPv4 a
func IsIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

// IsIPv6 .
func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

// MinMax https://stackoverflow.com/a/45976758/12985309
func MinMax(array []int64) (int64, int64) {
	var max int64 = array[0]
	var min int64 = array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}
