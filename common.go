package main

import (
	"errors"
	"strings"
)

// ParseStrIntoMap Takes a string in the format of  this=that;this=theother and
// Makes a map from it
func ParseStrIntoMap(in string) (map[string]string, error) {
	answer := make(map[string]string)
	atoms := strings.Split(in, ";")
	for _, atom := range atoms {
		subs := strings.SplitN(atom, "=", 2)
		if len(subs) < 2 {
			return nil, errors.New("Incorrect song format. Please ask me for help \"" + atom + "\"")
		}
		answer[strings.TrimSpace(subs[0])] = strings.TrimSpace(subs[1])
	}

	return answer, nil
}
