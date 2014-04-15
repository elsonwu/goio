package goreal

import (
	"sync"
)

type cs map[string]*Client

type Clients struct {
	cs   cs
	lock *sync.RWMutex
}

func (self *Clients) Get(id string) *Client {
	if clt, ok := self.cs[id]; ok {
		return clt
	}

	return nil
}

func (self *Clients) Count() int {
	return len(self.cs)
}

func (self *Clients) Receive(message *Message) {
	for _, clt := range self.cs {
		go func(clt *Client, msg *Message) {
			clt.Msg <- msg
		}(clt, message)
	}
}

func (self *Clients) Delete(id string) {
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.cs, id)
}

func (self *Clients) Add(clt *Client) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if nil != self.Get(clt.Id) {
		return
	}

	self.cs[clt.Id] = clt
	clt.On("destory", func(message *Message) {
		self.Delete(clt.Id)
	})
}
