package goio

import (
	"sync"
)

type MClients struct {
	clients []*Clients
	current *Clients
}

func (self *MClients) Init() {
	if self.current == nil || self.clients == nil {
		self.clients = NewClients()
	}

	if self.current == nil || 1000 < self.current.Count() {
		self.current = &Clients{
			Map: make(map[string]*Client),
		}
		self.clients = append(self.clients, self.current)
	}
}

func (self *MClients) Get(id string) *Client {
	for _, cs := range self.clients {
		if clt := cs.Get(id); clt != nil {
			return clt
		}
	}

	return nil
}

func (self *MClients) Count() int {
	c := 0
	for _, cs := range self.clients {
		c += cs.Count()
	}
	return c
}

func (self *MClients) Delete(id string) {
	for _, cs := range self.clients {
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

type Clients struct {
	Map  map[string]*Client
	lock sync.RWMutex
}

func (self *Clients) Get(id string) *Client {
	if clt, ok := self.Map[id]; ok {
		return clt
	}

	return nil
}

func (self *Clients) Count() int {
	return len(self.Map)
}

func (self *Clients) Delete(id string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.Map, id)
}

func (self *Clients) Add(clt *Client) {
	if nil != self.Get(clt.Id) {
		return
	}

	clt.On("destroy", func(message *Message) {
		self.Delete(clt.Id)
	})

	self.lock.Lock()
	defer self.lock.Unlock()
	self.Map[clt.Id] = clt
}
