package goio

import (
	"sync"
)

type Rooms struct {
	m    map[string]*Room
	lock sync.Mutex
}

func (self *Rooms) Count() int {
	return len(self.m)
}

func (self *Rooms) Delete(id string) {
	delete(self.m, id)
}

func (self *Rooms) Add(room *Room) {
	if self.Has(room.Id) {
		return
	}

	self.m[room.Id] = room
	room.On("broadcast", func(message *Message) {
		room.Receive(message)
	})

	room.On("destory", func(message *Message) {
		self.Delete(room.Id)
	})
}

func (self *Rooms) Has(id string) bool {
	_, ok := self.m[id]
	return ok
}

func (self *Rooms) Get(id string) *Room {
	if room, ok := self.m[id]; ok {
		return room
	}

	room := newRoom(id)
	GlobalRooms().Add(room)
	return room
}
