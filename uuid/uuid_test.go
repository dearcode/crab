package uuid

import (
	"testing"
)

func TestUUID(t *testing.T) {
	for i := 0; i < 1000; i++ {
		t.Logf("%s", String())
	}

}
