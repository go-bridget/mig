package db

import "regexp"

var dsnMasker = regexp.MustCompile("(.)(?:.*)(.):(.)(?:.*)(.)@")

func maskDSN(dsn string) string {
	return dsnMasker.ReplaceAllString(dsn, "$1****$2:$3****$4@")
}
