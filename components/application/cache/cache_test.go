package cache

import (
	"sync"
	"testing"

	extcache "github.com/go-zoox/cache"
)

func resetStateForTest() {
	once = sync.Once{}
	instance = nil
}

func shouldPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, but did not panic")
		}
	}()

	fn()
}

func TestGetPanicsWhenNotInitialized(t *testing.T) {
	resetStateForTest()

	shouldPanic(t, func() {
		_ = Get()
	})
}

func TestSetAndGetReturnSameInstance(t *testing.T) {
	resetStateForTest()

	c := extcache.New()
	Set(c)

	got := Get()
	if got != c {
		t.Fatal("Get() should return the same cache instance set by Set()")
	}
}

func TestSetPanicsWhenNil(t *testing.T) {
	resetStateForTest()

	shouldPanic(t, func() {
		Set(nil)
	})
}
