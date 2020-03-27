package skiplist

import (
	"fmt"
	"testing"
)

func TestInsert(t *testing.T) {
	table := New(3)
	table.Insert("aaaaaaaaaaa", 111111)
	fmt.Printf("---------%v\n", table)
}
