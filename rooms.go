package goio

import (
	"sync"
)

type Rooms struct {
	Map  map[string]*Room
	lock sync.RWMutex
}

func (self *Rooms) Count() int {
	self.lock.Lock()
	defer self.lock.Unlock()

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

	self.lock.Lock()
	defer self.lock.Unlock()

	room.On("destroy", func(message *Message) {
		self.Delete(room.Id)
	})

	self.Map[room.Id] = room
}

func (self *Rooms) Has(id string) bool {
	self.lock.Lock()
	defer self.lock.Unlock()

	_, ok := self.Map[id]
	return ok
}

func (self *Rooms) Get(id string, autoNew bool) *Room {
	self.lock.Lock()

	if room, ok := self.Map[id]; ok {
		self.lock.Unlock()
		return room
	}

	self.lock.Unlock()

	if autoNew {
		return NewRoom(id)
	}

	return nil
}
