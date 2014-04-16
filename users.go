package goio

import (
	"sync"
)

type Users struct {
	m    map[string]*User
	lock sync.Mutex
}

func (self *Users) Count() int {
	return len(self.m)
}

func (self *Users) Add(user *User) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if nil != self.Get(user.Id) {
		return
	}

	user.On("destory", func(message *Message) {
		self.Delete(user.Id)
	})

	self.m[user.Id] = user
}

func (self *Users) Delete(userId string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.m, userId)
}

func (self *Users) Get(userId string) *User {
	if user, ok := self.m[userId]; ok {
		return user
	}

	return nil
}
