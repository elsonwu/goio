package goio

import (
	"sync"
)

type Users struct {
	Map  map[string]*User
	lock sync.RWMutex
}

func (self *Users) Receive(message *Message) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, user := range self.Map {
		if user != nil {
			user.Receive(message)
		}
	}
}

func (self *Users) Each(callback func(*User)) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, user := range self.Map {
		callback(user)
	}
}

func (self *Users) Count() int {
	return len(self.Map)
}

func (self *Users) Add(user *User) {
	if nil != self.Get(user.Id) {
		return
	}

	user.On("destroy", func(message *Message) {
		self.Delete(user.Id)
	})

	// send global message to all members
	self.Receive(&Message{
		EventName: "join",
		CallerId:  user.Id,
	})

	self.lock.Lock()
	defer self.lock.Unlock()

	self.Map[user.Id] = user
}

func (self *Users) Delete(userId string) {
	self.lock.Lock()
	delete(self.Map, userId)
	self.lock.Unlock()

	// send global message to all members
	self.Receive(&Message{
		EventName: "leave",
		CallerId:  userId,
	})
}

func (self *Users) Get(userId string) *User {
	if user, ok := self.Map[userId]; ok {
		return user
	}

	return nil
}
