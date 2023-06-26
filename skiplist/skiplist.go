package skiplist

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"
)

type node struct {
	key   string
	value interface{}
	level int
	next  []*node
}

type nodes *node

type Skiplist struct {
	maxLevel int
	lists    []nodes
	rand     *rand.Rand
}

func New(maxLevel int) *Skiplist {

	lists := make([]nodes, maxLevel)
	first := &node{
		next: make([]*node, maxLevel),
	}
	for i := range lists {
		lists[i] = first
	}

	return &Skiplist{
		maxLevel: maxLevel,
		lists:    lists,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Skiplist) Insert(key string, value interface{}) {
	pns := s.priorNodes(key)
	level := s.randomLevel()

	n := &node{
		key:   key,
		value: value,
		level: level,
		next:  make([]*node, s.maxLevel),
	}

	for i := level; i < s.maxLevel; i++ {
		n.next[i] = pns[i].next[i]
		pns[i].next[i] = n
	}
}

func (s *Skiplist) Delete(key string) {

}

func (s *Skiplist) Get(key string) interface{} {
	ptr := s.lists[0]
	for i := range s.lists {
		for ; ptr != nil; ptr = ptr.next[i] {
			if ptr.key == key {
				return ptr.value
			}
			if ptr.next[i] == nil || key < ptr.next[i].key {
				break
			}
		}
	}

	return nil
}

func (s *Skiplist) randomLevel() int {
	return s.rand.Intn(s.maxLevel)
}

func (s *Skiplist) priorNodes(key string) []*node {
	pns := make([]*node, s.maxLevel)
	ptr := s.lists[0]
	for i := range s.lists {
		pns[i] = ptr
		for ; ptr != nil; ptr = ptr.next[i] {
			if key < ptr.key || ptr.next[i] == nil {
				pns[i] = ptr
				break
			}
		}
	}
	return pns
}

func (s *Skiplist) String() string {
	buf := bytes.NewBuffer(nil)

	for i := range s.lists {
		fmt.Fprintf(buf, "level[%v]:", i)
		nodes := s.lists[i]
		for ptr := nodes; ptr != nil; ptr = ptr.next[i] {
			fmt.Fprintf(buf, "key[%d]:%s\t", ptr.level, ptr.key)
		}
		fmt.Fprintf(buf, "\n")
	}
	return buf.String()
}
