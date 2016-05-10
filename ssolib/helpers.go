package ssolib

import (
	"regexp"
	"strings"
)

var splitRegexp = regexp.MustCompile(`\s+`)

// A split function which be able to handle multiple space delimeter, like python's split()
func split(s string) []string {
	return splitRegexp.Split(strings.Trim(s, " "), -1)
}
