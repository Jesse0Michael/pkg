package data

import (
	"maps"
	"slices"
	"sync"
	"testing"
)

func TestMap_SetAndGet(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("test-key", 42)

	v, ok := m.Get("test-key")
	if !ok {
		t.Fatal("Get returned ok=false, want true")
	}
	if v != 42 {
		t.Errorf("Get returned %d, want 42", v)
	}
}

func TestMap_GetMissing(t *testing.T) {
	m := NewMap[string, int]()

	v, ok := m.Get("test-key")
	if ok {
		t.Error("Get returned ok=true for missing key")
	}
	if v != 0 {
		t.Errorf("Get returned %d, want zero value", v)
	}
}

func TestMap_Delete(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("test-key", 1)
	m.Delete("test-key")

	if _, ok := m.Get("test-key"); ok {
		t.Error("key still present after Delete")
	}
}

func TestMap_Len(t *testing.T) {
	m := NewMap[string, int]()
	if got := m.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}

	m.Set("test-key", 1)
	if got := m.Len(); got != 1 {
		t.Errorf("Len() = %d, want 1", got)
	}
}

func TestMap_Clear(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("test-key-1", 1)
	m.Set("test-key-2", 2)
	m.Clear()
	if got := m.Len(); got != 0 {
		t.Errorf("Len() = %d after Clear, want 0", got)
	}
}

func TestMap_All(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("test-key-1", 1)
	m.Set("test-key-2", 2)
	m.Set("test-key-3", 3)

	want := map[string]int{"test-key-1": 1, "test-key-2": 2, "test-key-3": 3}
	got := maps.Collect(m.All())
	if !maps.Equal(got, want) {
		t.Errorf("All() = %v, want %v", got, want)
	}
}

func TestMap_All_EarlyStop(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("test-key-1", 1)
	m.Set("test-key-2", 2)

	count := 0
	for range m.All() {
		count++
		break
	}
	if count != 1 {
		t.Errorf("iterated %d times after break, want 1", count)
	}
}

func TestMap_Keys(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("test-key-1", 1)
	m.Set("test-key-2", 2)

	keys := slices.Sorted(m.Keys())
	want := []string{"test-key-1", "test-key-2"}
	if !slices.Equal(keys, want) {
		t.Errorf("Keys() = %v, want %v", keys, want)
	}
}

func TestMap_Values(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("test-key-1", 1)
	m.Set("test-key-2", 2)

	values := slices.Sorted(m.Values())
	want := []int{1, 2}
	if !slices.Equal(values, want) {
		t.Errorf("Values() = %v, want %v", values, want)
	}
}

func TestMap_ConcurrentAccess(t *testing.T) {
	m := NewMap[int, int]()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Set(i, i*2)
			m.Get(i)
		}()
	}
	wg.Wait()

	if got := m.Len(); got != 100 {
		t.Errorf("Len() = %d, want 100", got)
	}
}
