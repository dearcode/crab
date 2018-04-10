package cache

import (
	"testing"
	"time"
)

func TestCacheActive(t *testing.T) {
	data := struct {
		User string
		Pass string
	}{
		"test@jd.com",
		"password",
	}
	c := NewCache(2)

	c.Add("1", &data)
	c.Get("1")
	time.Sleep(time.Second)
	val := c.Get("1")
	if val == nil {
		t.Fatalf("not found, expect %v", data)
	}
}

func TestCacheInactive(t *testing.T) {
	data := struct {
		User string
		Pass string
	}{
		"test@jd.com",
		"password",
	}
	c := NewCache(2)

	c.Add("1", &data)
	val := c.Get("1")

	time.Sleep(time.Second * 2)

	c.Get("1")
	if val != nil {
		t.Fatalf("expect not found")
	}
}
