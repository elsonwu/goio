package goreal

import (
	"sync"
)

type rs map[string]*Room

type Rooms struct {
	rs   rs
	lock *sync.RWMutex
}

func (self *Rooms) Count() int {
	return len(self.rs)
}

func (self *Rooms) Delete(id string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.rs, id)
}

func (self *Rooms) Add(room *Room) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if self.Has(room.Id) {
		return
	}

	self.rs[room.Id] = room

	room.On("broadcast", func(message *Message) {
		room.Receive(message)
	})

	room.On("destory", func(message *Message) {
		self.Delete(room.Id)
	})

	room.On("join", func(message *Message) {
		room.Receive(message)
	})

	room.On("leave", func(message *Message) {
		room.Receive(message)
	})
}

func (self *Rooms) Has(id string) bool {
	_, ok := self.rs[id]
	return ok
}

func (self *Rooms) Get(id string) *Room {
	if room, ok := self.rs[id]; ok {
		return room
	}

	self.lock.Lock()
	defer self.lock.Unlock()

	room := newRoom(id)
	GlobalRooms().Add(room)
	return room
}
