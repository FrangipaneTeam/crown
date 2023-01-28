package common

import (
	"regexp"
)

// ReSubMatchMap returns a map of the submatches of the regexp r in the string str.
func ReSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	if len(match) == 0 {
		return subMatchMap
	}
	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	return subMatchMap
}
