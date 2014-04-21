package goio

import (
	"log"
	"sync"
)

type User struct {
	Event
	Id        string
	ClientIds MapBool
	RoomIds   MapBool
	lock      sync.RWMutex
}

func (self *User) Receive(message *Message) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for cltId := range self.ClientIds.Map {
		clt := GlobalClients().Get(cltId)
		if clt != nil {
			clt.Receive(message)
		}
	}
}

func (self *User) Delete(id string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.ClientIds.Delete(id)
	if 0 == self.ClientIds.Count() {
		log.Println("user client count 0")
		self.Destory()
	}
}

func (self *User) Destory() {
	self.Emit("destory", nil)
}

func (self *User) Add(clt *Client) {
	if self.ClientIds.Has(clt.Id) {
		return
	}

	clt.On("destory", func(message *Message) {
		self.Delete(clt.Id)
	})

	clt.UserId = self.Id
	self.ClientIds.Add(clt.Id)
}
