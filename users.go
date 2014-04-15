package goreal

import (
	"sync"
)

type us map[string]*User

type Users struct {
	us   us
	lock *sync.RWMutex
}

func (self *Users) Count() int {
	return len(self.us)
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

	self.us[user.Id] = user
}

func (self *Users) Delete(userId string) {
	self.lock.Lock()
	defer self.lock.Unlock()

	delete(self.us, userId)
}

func (self *Users) Get(userId string) *User {
	if user, ok := self.us[userId]; ok {
		return user
	}

	return nil
}
