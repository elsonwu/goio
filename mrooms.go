package goio

import (
	"sync"
)

type MRooms struct {
	Rooms   []*Rooms
	current *Rooms
	max     int
	lock    sync.RWMutex
}

func (self *MRooms) Init() {
	if self.current == nil || self.Rooms == nil {
		self.Rooms = NewRooms()
	}

	if self.current == nil || self.max < self.current.Count() {
		self.current = &Rooms{
			Map: make(map[string]*Room),
		}
		self.Rooms = append(self.Rooms, self.current)
	}
}

func (self *MRooms) Each(callback func(r *Room)) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, cs := range self.Rooms {
		cs.Each(callback)
	}
}

func (self *MRooms) Get(id string, autoNew bool) *Room {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, cs := range self.Rooms {
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
	self.lock.RLock()
	defer self.lock.RUnlock()

	c := 0
	for _, cs := range self.Rooms {
		c += cs.Count()
	}
	return c
}

func (self *MRooms) Delete(id string) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, cs := range self.Rooms {
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
