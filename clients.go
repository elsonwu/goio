package goio

import (
	"sync"
)

type Clients struct {
	m    map[string]*Client
	lock sync.Mutex
}

func (self *Clients) Get(id string) *Client {
	if clt, ok := self.m[id]; ok {
		return clt
	}

	return nil
}

func (self *Clients) Count() int {
	return len(self.m)
}

func (self *Clients) Receive(message *Message) {
	for _, clt := range self.m {
		go func(clt *Client, msg *Message) {
			clt.Msg <- msg
		}(clt, message)
	}
}

func (self *Clients) Delete(id string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.m, id)
}

func (self *Clients) Add(clt *Client) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if nil != self.Get(clt.Id) {
		return
	}

	self.m[clt.Id] = clt
	clt.On("destory", func(message *Message) {
		self.Delete(clt.Id)
	})
}
