package goio

import (
	"sync"
)

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

	self.lock.Lock()
	defer self.lock.Unlock()
	clt.On("destroy", func(message *Message) {
		self.Delete(clt.Id)
	})

	self.Map[clt.Id] = clt
}
