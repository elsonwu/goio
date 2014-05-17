package goio

import (
	"sync"
)

type Clients struct {
	Map  map[string]*Client
	lock sync.RWMutex
}

func (self *Clients) Each(callback func(*Client)) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, clt := range self.Map {
		callback(clt)
	}
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
