package httpapi

import "regexp"

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// validKey prüft, ob k nicht leer ist, nur aus [A-Za-z0-9._-] besteht
// und höchstens 128 Zeichen lang ist.
func validKey(k string) bool {
	return k != "" && len(k) <= 128 && keyPattern.MatchString(k)
}

// validDescription prüft, ob d höchstens 2048 Zeichen lang ist.
func validDescription(d string) bool {
	return len(d) <= 2048
}

// validRolloutPercent prüft, ob p im Bereich 0–100 liegt.
func validRolloutPercent(p int) bool {
	return p >= 0 && p <= 100
}
