package goio

import (
	"sync"
)

type Users struct {
	Map  map[string]*User
	lock sync.RWMutex
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

	self.lock.Lock()
	defer self.lock.Unlock()
	user.On("destroy", func(message *Message) {
		self.Delete(user.Id)
	})

	self.Map[user.Id] = user
}

func (self *Users) Delete(userId string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.Map, userId)
}

func (self *Users) Get(userId string) *User {
	self.lock.Lock()
	defer self.lock.Unlock()

	if user, ok := self.Map[userId]; ok {
		return user
	}

	return nil
}
