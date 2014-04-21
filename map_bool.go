package goio

import (
	"sync"
)

type mapBool map[string]bool

type MapBool struct {
	Map     mapBool
	Max     int
	IsFull  bool
	IsEmpty bool
	lock    sync.RWMutex
}

func (self *MapBool) Add(str string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.Map[str] = true
	if self.Max <= len(self.Map) {
		self.IsFull = true
		self.IsEmpty = false
	}
}

func (self *MapBool) Delete(str string) bool {
	self.lock.Lock()
	defer self.lock.Unlock()

	if !self.Has(str) {
		return false
	}

	delete(self.Map, str)
	if 0 == len(self.Map) {
		self.IsEmpty = true
		self.IsFull = false
	}

	return true
}

func (self *MapBool) Has(id string) bool {
	_, ok := self.Map[id]
	return ok
}

func (self *MapBool) Count() int {
	return len(self.Map)
}
