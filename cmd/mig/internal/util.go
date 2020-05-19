package internal

import (
	"strings"

	"github.com/serenize/snaker"
)

func Camel(input string) string {
	if strings.ToLower(input) == "id" {
		return "ID"
	}
	// special case from having camel case `showId` fields in DB
	if len(input) > 2 && input[len(input)-2:] == "Id" {
		input = input[0:len(input)-2] + "_id"
	}
	return snaker.SnakeToCamel(input)
}

func Contains(set []string, value string) bool {
	for _, v := range set {
		if v == value {
			return true
		}
	}
	return false
}
