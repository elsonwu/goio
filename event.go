package goio

import (
	"sync"
)

type callback func(message *Message)

type Event struct {
	evts map[string][]callback
	lock sync.RWMutex
}

func (self *Event) On(eventName string, fn callback) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if self.evts == nil {
		self.evts = make(map[string][]callback)
	}

	if _, ok := self.evts[eventName]; !ok {
		self.evts[eventName] = make([]callback, 0)
	}

	self.evts[eventName] = append(self.evts[eventName], fn)
}

func (self *Event) Emit(eventName string, message *Message) {
	self.lock.RLock()
	evts, ok := self.evts[eventName]
	self.lock.RUnlock()

	if !ok {
		return
	}

	for _, fn := range evts {
		fn(message)
	}
}
