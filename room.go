package goio

import (
	"sync"
)

type Room struct {
	Event
	Id      string
	UserIds MapBool
	lock    sync.RWMutex
}

func (self *Room) Has(id string) bool {
	return self.UserIds.Has(id)
}

func (self *Room) Receive(message *Message) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for uid := range self.UserIds.Map {
		user := GlobalUsers().Get(uid)
		if user != nil {
			user.Receive(message)
		}
	}
}

func (self *Room) Delete(id string) {
	self.UserIds.Delete(id)
	self.Receive(&Message{
		EventName: "leave",
		CallerId:  id,
		RoomId:    self.Id,
	})

	if 0 == self.UserIds.Count() {
		self.Destory()
	}
}

func (self *Room) Destory() {
	self.Emit("destory", nil)
}

func (self *Room) Add(user *User) {
	if self.Has(user.Id) {
		return
	}

	self.Receive(&Message{
		EventName: "join",
		RoomId:    self.Id,
		CallerId:  user.Id,
	})

	user.On("destory", func(message *Message) {
		self.Delete(user.Id)
	})

	user.RoomIds.Add(self.Id)
	self.UserIds.Add(user.Id)
}
