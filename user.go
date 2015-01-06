package goio

import (
	"log"
	// "sync"
)

type User struct {
	Event
	Id        string
	ClientIds MapBool
	RoomIds   MapBool
	data      *TempData
}

func (self *User) Data() *TempData {
	if self.data == nil {
		self.data = &TempData{}
	}

	return self.data
}

func (self *User) Receive(message *Message) {
	self.ClientIds.Each(func(cltId string) {

		// don't send message to the client itself
		if message.ClientId == cltId {
			return
		}

		if clt := GlobalClients().Get(cltId); clt != nil {
			clt.Receive(message)
		}
	})
}

func (self *User) Delete(id string) {
	self.ClientIds.Delete(id)

	if 0 == self.ClientIds.Count() {
		if Debug {
			log.Println("user client count 0")
		}

		self.Destroy()
	}
}

func (self *User) Destroy() {
	self.Emit("destroy", nil)
}

func (self *User) Add(clt *Client) {
	if self.ClientIds.Has(clt.Id) {
		return
	}

	clt.On("destroy", func(message *Message) {
		self.Delete(clt.Id)
	})

	clt.UserId = self.Id
	self.ClientIds.Add(clt.Id)
}
