package goio

import (
	"sync"
)

type MRooms struct {
	rooms   []*Rooms
	current *Rooms
}

func (self *MRooms) Init() {
	if self.current == nil || self.rooms == nil {
		self.rooms = NewRooms()
	}

	if self.current == nil || 100 < self.current.Count() {
		self.current = &Rooms{
			Map: make(map[string]*Room),
		}
		self.rooms = append(self.rooms, self.current)
	}
}

func (self *MRooms) Each(callback func(r *Room)) {
	for _, cs := range self.rooms {
		cs.Each(callback)
	}
}

func (self *MRooms) Get(id string, autoNew bool) *Room {
	for _, cs := range self.rooms {
		if clt := cs.Get(id, false); clt != nil {
			return clt
		}
	}

	if autoNew {
		self.Init()
		return self.current.Get(id, autoNew)
	}

	return nil
}

func (self *MRooms) Count() int {
	c := 0
	for _, cs := range self.rooms {
		c += cs.Count()
	}
	return c
}

func (self *MRooms) Delete(id string) {
	for _, cs := range self.rooms {
		if cs.Has(id) {
			cs.Delete(id)
			return
		}
	}
}

func (self *MRooms) Add(clt *Room) {
	if nil == self.Get(clt.Id, false) {
		self.Init()
		self.current.Add(clt)
	}
}

type Rooms struct {
	Map  map[string]*Room
	lock sync.RWMutex
}

func (self *Rooms) Each(callback func(*Room)) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, r := range self.Map {
		callback(r)
	}
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
	room, ok := self.Map[id]
	if ok {
		return room
	}

	if autoNew {
		return NewRoom(id)
	}

	return nil
}
