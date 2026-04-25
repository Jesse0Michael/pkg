package data

import (
	"maps"
	"slices"
	"testing"
)

func TestDict_SetAndGet(t *testing.T) {
	d := NewDict[int]()
	d.Set("test-key", 42)

	v, ok := d.Get("test-key")
	if !ok {
		t.Fatal("Get returned ok=false, want true")
	}
	if v != 42 {
		t.Errorf("Get returned %d, want 42", v)
	}
}

func TestDict_GetMissing(t *testing.T) {
	d := NewDict[int]()

	v, ok := d.Get("test-key")
	if ok {
		t.Error("Get returned ok=true for missing key")
	}
	if v != 0 {
		t.Errorf("Get returned %d, want zero value", v)
	}
}

func TestDict_Delete(t *testing.T) {
	d := NewDict[int]()
	d.Set("test-key", 1)
	d.Delete("test-key")

	if _, ok := d.Get("test-key"); ok {
		t.Error("key still present after Delete")
	}
}

func TestDict_Len(t *testing.T) {
	d := NewDict[int]()
	if got := d.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}

	d.Set("test-key", 1)
	if got := d.Len(); got != 1 {
		t.Errorf("Len() = %d, want 1", got)
	}
}

func TestDict_All(t *testing.T) {
	d := NewDict[int]()
	d.Set("test-key-1", 1)
	d.Set("test-key-2", 2)

	want := map[string]int{"test-key-1": 1, "test-key-2": 2}
	got := maps.Collect(d.All())
	if !maps.Equal(got, want) {
		t.Errorf("All() = %v, want %v", got, want)
	}
}

func TestDict_Keys(t *testing.T) {
	d := NewDict[int]()
	d.Set("test-key-1", 1)
	d.Set("test-key-2", 2)

	keys := slices.Sorted(d.Keys())
	want := []string{"test-key-1", "test-key-2"}
	if !slices.Equal(keys, want) {
		t.Errorf("Keys() = %v, want %v", keys, want)
	}
}

func TestDict_Values(t *testing.T) {
	d := NewDict[int]()
	d.Set("test-key-1", 1)
	d.Set("test-key-2", 2)

	values := slices.Sorted(d.Values())
	want := []int{1, 2}
	if !slices.Equal(values, want) {
		t.Errorf("Values() = %v, want %v", values, want)
	}
}

func TestDict_AssignableToMap(t *testing.T) {
	// Dict is a type alias, so it should be assignable to *Map[string, V].
	var m *Map[string, int] = NewDict[int]()
	m.Set("test-key", 1)

	if v, ok := m.Get("test-key"); !ok || v != 1 {
		t.Errorf("Map assignment failed: got (%d, %v)", v, ok)
	}
}
