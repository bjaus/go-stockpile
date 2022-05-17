package stockpile

import (
	"fmt"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	c := New(0)

	_, ok := c.Get("a")
	if ok {
		t.Errorf("got %t, want %t", ok, false)
	}

	c.Set("b", 123, 0)
	_, ok = c.Get("b")
	if !ok {
		t.Errorf("got %t, want %t", ok, true)
	}

	c.SetNoExpiry("c", 123)
	_, ok = c.Get("c")
	if !ok {
		t.Errorf("got %t, want %t", ok, true)
	}

	c.Delete("d")
	_, ok = c.Get("d")
	if ok {
		t.Errorf("got %t, want %t", ok, false)
	}

	c.Set("e", 123, 0)
	c.Reset()
	_, ok = c.Get("e")
	if ok {
		t.Errorf("got %t, want %t", ok, false)
	}
}

func TestCache_Count(t *testing.T) {
	c := New(0)

	count := c.Count()
	if count != 0 {
		t.Errorf("new cache count: should have zero items")
	}

	c.Set("a", "test", 0)
	count = c.Count()
	if count != 1 {
		t.Errorf("one item count: got %d, want %d", count, 1)
	}

	c.Reset()
	count = c.Count()
	if count != 0 {
		t.Errorf("after reset count: got %d, want %d", count, 0)
	}

	num := 10
	for i := 0; i < num; i++ {
		c.Set(fmt.Sprintf("%d", i), i, 0)
	}
	count = c.Count()
	if count != num {
		t.Errorf("multiple item count: got %d, want %d", count, num)
	}
}

func TestCacheExpiration(t *testing.T) {
	c := New(time.Millisecond * 1)

	k1 := "a"
	k2 := "b"
	k3 := "c"

	c.SetNoExpiry(k1, 1)
	c.Set(k2, 2, time.Millisecond*10)
	c.Set(k3, 3, time.Millisecond*50)

	<-time.After(time.Millisecond * 11)
	if _, ok := c.Get(k2); ok {
		t.Errorf("%q: should have expired", k2)
	}

	if _, ok := c.Get(k3); !ok {
		t.Errorf("%q: should not have expired", k3)
	}

	<-time.After(time.Millisecond * 40)
	if _, ok := c.Get(k3); ok {
		t.Errorf("%q: should have expired", k3)
	}

	if _, ok := c.Get(k1); !ok {
		t.Errorf("%q: should not expire", k1)
	}
}
