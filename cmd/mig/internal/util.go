package internal

import (
	"strings"

	stylecheck "honnef.co/go/tools/config"
)

// Camel converts a snake_case string to CamelCase.
func Camel(input string) string {
	// special case from having camel case `showId` fields in DB
	if len(input) > 2 && input[len(input)-2:] == "Id" {
		input = input[0:len(input)-2] + "_id"
	}

	// split string and check against initialisms
	keys := strings.Split(input, "_")
	for k, v := range keys {
		upper := strings.ToUpper(v)
		if Contains(stylecheck.DefaultConfig.Initialisms, upper) {
			keys[k] = upper
			continue
		}
		keys[k] = upper[0:1] + v[1:]
	}

	return strings.Join(keys, "")
}

// Title converts a snake_case string to Title Case with spaces.
func Title(input string) string {
	// special case from having camel case `showId` fields in DB
	if len(input) > 2 && input[len(input)-2:] == "Id" {
		input = input[0:len(input)-2] + "_id"
	}

	// split string and check against initialisms
	keys := strings.Split(input, "_")
	for k, v := range keys {
		upper := strings.ToUpper(v)
		if Contains(stylecheck.DefaultConfig.Initialisms, upper) {
			keys[k] = upper
			continue
		}
		keys[k] = upper[0:1] + v[1:]
	}

	return strings.Join(keys, " ")
}

// Filename converts a string to a lowercase filename with underscores.
func Filename(input string) string {
	return strings.ReplaceAll(strings.ToLower(input), " ", "_")
}

// Contains checks if a value exists in a string slice.
func Contains(set []string, value string) bool {
	for _, v := range set {
		if v == value {
			return true
		}
	}
	return false
}
