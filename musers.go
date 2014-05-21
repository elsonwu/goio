package goio

import (
	"sync"
)

type MUsers struct {
	Users   []*Users
	current *Users
	max     int
	lock    sync.RWMutex
}

func (self *MUsers) Init() {
	if self.current == nil || self.Users == nil {
		self.Users = NewUsers()
	}

	if self.current == nil || self.max < self.current.Count() {
		self.current = &Users{
			Map: make(map[string]*User),
		}
		self.Users = append(self.Users, self.current)
	}
}

func (self *MUsers) Each(callback func(*User)) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, u := range self.Users {
		u.Each(callback)
	}
}

func (self *MUsers) Get(id string) *User {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, cs := range self.Users {
		if clt := cs.Get(id); clt != nil {
			return clt
		}
	}

	return nil
}

func (self *MUsers) Count() int {
	self.lock.RLock()
	defer self.lock.RUnlock()

	c := 0
	for _, cs := range self.Users {
		c += cs.Count()
	}
	return c
}

func (self *MUsers) Delete(id string) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	for _, cs := range self.Users {
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
