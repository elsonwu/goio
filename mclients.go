package goio

import (
	"sync"
)

type MClients struct {
	Clients []*Clients
	current *Clients
	max     int
	lock    sync.RWMutex
}

func (self *MClients) Init() {
	if self.current == nil || self.Clients == nil {
		self.Clients = NewClients()
	}

	if self.current == nil || self.max < self.current.Count() {
		self.current = &Clients{
			Map: make(map[string]*Client),
		}
		self.Clients = append(self.Clients, self.current)
	}
}

func (self *MClients) Each(callback func(*Client)) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, cs := range self.Clients {
		cs.Each(callback)
	}
}

func (self *MClients) Get(id string) *Client {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, cs := range self.Clients {
		if clt := cs.Get(id); clt != nil {
			return clt
		}
	}

	return nil
}

func (self *MClients) Count() int {
	self.lock.RLock()
	defer self.lock.RUnlock()

	c := 0
	for _, cs := range self.Clients {
		c += cs.Count()
	}
	return c
}

func (self *MClients) Delete(id string) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, cs := range self.Clients {
		if clt := cs.Get(id); clt != nil {
			cs.Delete(id)
			return
		}
	}
}

func (self *MClients) Add(clt *Client) {
	if nil == self.Get(clt.Id) {
		self.Init()
		self.current.Add(clt)
	}
}
