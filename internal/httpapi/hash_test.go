package httpapi

import "testing"

func TestStableHashDeterministic(t *testing.T) {
	a := stableHash("flag.key", "user-1")
	b := stableHash("flag.key", "user-1")
	if a != b {
		t.Fatalf("expected deterministic hash, got %d and %d", a, b)
	}
}

func TestStableHashDistinguishesUser(t *testing.T) {
	a := stableHash("flag.key", "user-1")
	b := stableHash("flag.key", "user-2")
	if a == b {
		t.Fatalf("expected different users to produce different hashes, both %d", a)
	}
}

func TestStableHashDistinguishesKey(t *testing.T) {
	a := stableHash("flag.a", "user-1")
	b := stableHash("flag.b", "user-1")
	if a == b {
		t.Fatalf("expected different keys to produce different hashes, both %d", a)
	}
}

func TestStableHashSeparatesKeyAndUser(t *testing.T) {
	a := stableHash("ab", "c")
	b := stableHash("a", "bc")
	if a == b {
		t.Fatalf("expected key/user separation to avoid collision, both %d", a)
	}
}
