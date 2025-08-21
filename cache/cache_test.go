package cache

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func TestNewCache(t *testing.T) {
	s, _ := miniredis.Run()
	rc := redis.NewClient(&redis.Options{Addr: s.Addr()})
	got := NewCache(Config{Size: 100}, rc)
	if got == nil {
		t.Errorf("NewCache() = %v, want &Cache{}", got)
	}
}
