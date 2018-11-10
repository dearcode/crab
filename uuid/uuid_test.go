package uuid

import (
	"testing"
)

func TestUUID(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id := String()
		info, _ := Info(id)
		t.Logf("%s, %s", id, info)
	}
}
