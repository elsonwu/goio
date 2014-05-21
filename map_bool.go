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

func (self *MapBool) Each(callback func(string)) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for str, _ := range self.Map {
		callback(str)
	}
}

func (self *MapBool) Array() []string {
	arr := make([]string, 0, len(self.Map))

	self.lock.RLock()
	defer self.lock.RUnlock()

	for str, _ := range self.Map {
		arr = append(arr, str)
	}

	return arr
}

func (self *MapBool) Add(str string) {
	self.lock.Lock()
	self.Map[str] = true
	self.lock.Unlock()

	if self.Max <= len(self.Map) {
		self.IsFull = true
		self.IsEmpty = false
	}
}

func (self *MapBool) Delete(str string) bool {
	if !self.Has(str) {
		return false
	}

	self.lock.Lock()
	delete(self.Map, str)
	self.lock.Unlock()

	if 0 == len(self.Map) {
		self.IsEmpty = true
		self.IsFull = false
	}

	return true
}

func (self *MapBool) Has(id string) bool {
	self.lock.RLock()
	defer self.lock.RUnlock()

	_, ok := self.Map[id]
	return ok
}

func (self *MapBool) Count() int {
	return len(self.Map)
}
