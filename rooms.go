package goio

import (
	"sync"
)

type Rooms struct {
	Map  map[string]*Room
	lock sync.RWMutex
}

func (self *Rooms) Count() int {
	return len(self.Map)
}

func (self *Rooms) Delete(id string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.Map, id)
}

func (self *Rooms) Add(room *Room) {
	if self.Has(room.Id) {
		return
	}

	room.On("destroy", func(message *Message) {
		self.Delete(room.Id)
	})

	self.lock.Lock()
	defer self.lock.Unlock()

	self.Map[room.Id] = room
}

func (self *Rooms) Has(id string) bool {
	self.lock.RLock()
	defer self.lock.RUnlock()

	_, ok := self.Map[id]
	return ok
}

func (self *Rooms) Get(id string, autoNew bool) *Room {
	self.lock.RLock()
	room, ok := self.Map[id]
	self.lock.RUnlock()

	if ok {
		return room
	}

	if autoNew {
		return NewRoom(id)
	}

	return nil
}
