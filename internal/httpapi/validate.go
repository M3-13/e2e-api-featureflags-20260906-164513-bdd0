package httpapi

import "regexp"

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// validKey prüft, ob k nicht leer ist und nur aus [A-Za-z0-9._-] besteht.
func validKey(k string) bool {
	return k != "" && keyPattern.MatchString(k)
}

// validRolloutPercent prüft, ob p im Bereich 0–100 liegt.
func validRolloutPercent(p int) bool {
	return p >= 0 && p <= 100
}
