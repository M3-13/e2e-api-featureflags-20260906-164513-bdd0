package store

import (
	"fmt"
	"sync"
	"testing"
)

func TestCreateAndGet(t *testing.T) {
	s := NewStore(10)
	f := Flag{Key: "a", Enabled: true, Description: "d", RolloutPercent: 50}
	if err := s.Create(f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := s.Get("a")
	if !ok {
		t.Fatal("expected flag to exist")
	}
	if got != f {
		t.Fatalf("got %+v, want %+v", got, f)
	}
}

func TestCreateDuplicate(t *testing.T) {
	s := NewStore(10)
	f := Flag{Key: "a"}
	if err := s.Create(f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Create(f); err != ErrDuplicate {
		t.Fatalf("expected ErrDuplicate, got %v", err)
	}
}

func TestCreateLimit(t *testing.T) {
	s := NewStore(2)
	_ = s.Create(Flag{Key: "a"})
	_ = s.Create(Flag{Key: "b"})
	if err := s.Create(Flag{Key: "c"}); err != ErrLimit {
		t.Fatalf("expected ErrLimit, got %v", err)
	}
}

func TestList(t *testing.T) {
	s := NewStore(10)
	if got := s.List(); len(got) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(got))
	}
	_ = s.Create(Flag{Key: "a"})
	_ = s.Create(Flag{Key: "b"})
	if got := s.List(); len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestUpdate(t *testing.T) {
	s := NewStore(10)
	_ = s.Create(Flag{Key: "a", Enabled: false})

	if !s.Update("a", Flag{Enabled: true, RolloutPercent: 30}) {
		t.Fatal("expected update of existing flag to succeed")
	}
	got, ok := s.Get("a")
	if !ok {
		t.Fatal("expected flag to exist after update")
	}
	if !got.Enabled || got.RolloutPercent != 30 || got.Key != "a" {
		t.Fatalf("unexpected flag after update: %+v", got)
	}

	if s.Update("missing", Flag{}) {
		t.Fatal("expected update of missing flag to fail")
	}
}

func TestDelete(t *testing.T) {
	s := NewStore(10)
	_ = s.Create(Flag{Key: "a"})

	if !s.Delete("a") {
		t.Fatal("expected delete of existing flag to succeed")
	}
	if _, ok := s.Get("a"); ok {
		t.Fatal("expected flag to be removed")
	}
	if s.Delete("a") {
		t.Fatal("expected delete of missing flag to fail")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := NewStore(10000)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("k%d", i)
			_ = s.Create(Flag{Key: key})
			_, _ = s.Get(key)
			_ = s.Update(key, Flag{Key: key, Enabled: true})
			_ = s.List()
			_ = s.Delete(key)
		}(i)
	}
	wg.Wait()

	if got := s.List(); len(got) != 0 {
		t.Fatalf("expected empty store after concurrent create/delete, got %d entries", len(got))
	}
}
