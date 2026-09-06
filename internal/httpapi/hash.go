package httpapi

import "hash/fnv"

// stableHash berechnet einen deterministischen FNV-1a-Hash über key und user.
// key und user werden durch ein Null-Byte getrennt, damit unterschiedliche
// Kombinationen nicht kollidieren (z. B. key="ab", user="c" vs. key="a",
// user="bc").
func stableHash(key, user string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(user))
	return h.Sum64()
}
