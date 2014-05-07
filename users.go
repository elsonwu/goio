package goio

import (
	"sync"
)

type Users struct {
	Map  map[string]*User
	lock sync.RWMutex
}

func (self *Users) Receive(message *Message) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for _, user := range self.Map {
		if user != nil {
			user.Receive(message)
		}
	}
}

func (self *Users) Count() int {
	self.lock.Lock()
	defer self.lock.Unlock()

	return len(self.Map)
}

func (self *Users) Add(user *User) {
	if nil != self.Get(user.Id) {
		return
	}

	user.On("destroy", func(message *Message) {
		self.Delete(user.Id)
	})

	self.Receive(&Message{
		EventName: "join",
		CallerId:  user.Id,
	})

	self.lock.Lock()
	self.Map[user.Id] = user
	self.lock.Unlock()
}

func (self *Users) Delete(userId string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.Map, userId)

	// send global message to all members
	self.Receive(&Message{
		EventName: "leave",
		CallerId:  userId,
	})
}

func (self *Users) Get(userId string) *User {
	self.lock.Lock()
	defer self.lock.Unlock()

	if user, ok := self.Map[userId]; ok {
		return user
	}

	return nil
}
