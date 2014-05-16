package goio

import (
	"sync"
)

type MUsers struct {
	users   []*Users
	current *Users
}

func (self *MUsers) Init() {
	if self.current == nil || self.users == nil {
		self.users = NewUsers()
	}

	if self.current == nil || 1000 < self.current.Count() {
		self.current = &Users{
			Map: make(map[string]*User),
		}
		self.users = append(self.users, self.current)
	}
}

func (self *MUsers) Get(id string) *User {
	for _, cs := range self.users {
		if clt := cs.Get(id); clt != nil {
			return clt
		}
	}

	return nil
}

func (self *MUsers) Count() int {
	c := 0
	for _, cs := range self.users {
		c += cs.Count()
	}
	return c
}

func (self *MUsers) Delete(id string) {
	for _, cs := range self.users {
		if clt := cs.Get(id); clt != nil {
			cs.Delete(id)
			return
		}
	}
}

func (self *MUsers) Add(clt *User) {
	if nil == self.Get(clt.Id) {
		self.Init()
		self.current.Add(clt)
	}
}

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
