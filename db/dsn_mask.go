package db

import "regexp"

var dsnMasker = regexp.MustCompile("(.)(?:.*)(.):(.)(?:.*)(.)@")

func maskDSN(dsn string, driver string) string {
	if driver == "sqlite" {
		return dsn
	}
	return dsnMasker.ReplaceAllString(dsn, "$1****$2:$3****$4@")
}
