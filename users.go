package goio

import (
	"sync"

	"github.com/golang/glog"
)

func NewUsers() *users {
	us := new(users)
	return us
}

type users struct {
	m     sync.Map
	count int
}

func (us *users) addMessage(msg *Message) {
	glog.V(1).Infoln("user(s) message " + msg.EventName)
	us.m.Range(func(k interface{}, v interface{}) bool {
		u := v.(*User)
		if u == nil || u.IsDead() || u.Id == msg.CallerId {
			return true
		}

		u.addMessage(msg)
		return true
	})
}

func (us *users) addUser(u *User) {
	us.count += 1
	us.m.Store(u.Id, u)
}

func (us *users) delUser(u *User) {
	us.count -= 1
	us.m.Delete(u.Id)
}

func (us *users) Count() int {
	return us.count
}

func (us *users) Get(userId string) *User {
	v, ok := us.m.Load(userId)
	if !ok {
		return nil
	}

	return v.(*User)
}

func (us *users) MustGet(userId string) *User {
	u := us.Get(userId)
	if u == nil {
		u = newUser(userId)
	}

	return u
}

func (us *users) Range(f func(r *User)) {
	us.m.Range(func(k interface{}, v interface{}) bool {
		user, ok := v.(*User)
		if !ok {
			return true
		}

		f(user)
		return true
	})
}
