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
	self.lock.RLock()
	defer self.lock.RUnlock()

	for uid := range self.UserIds.Map {
		if user := GlobalUsers().Get(uid); user != nil {
			user.Receive(message)
		}
	}
}

func (self *Room) Delete(userId string) {
	self.UserIds.Delete(userId)
	user := GlobalUsers().Get(userId)
	if user != nil {
		user.RoomIds.Delete(self.Id)
	}

	self.Receive(&Message{
		EventName: "leave",
		CallerId:  userId,
		RoomId:    self.Id,
	})

	if 0 == self.UserIds.Count() {
		self.Destroy()
	}
}

func (self *Room) Destroy() {
	self.Emit("destroy", nil)
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

	user.On("destroy", func(message *Message) {
		self.Delete(user.Id)
	})

	user.RoomIds.Add(self.Id)
	self.UserIds.Add(user.Id)
}
