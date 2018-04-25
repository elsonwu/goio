package goio

import (
	"sync"
)

func NewUsers() *users {
	us := new(users)
	return us
}

type users struct {
	m sync.Map
}

func (us *users) addUser(u *User) {
	us.m.Store(u.Id, u)
}

func (us *users) delUser(userId string) {
	us.m.Delete(userId)
}

func (us *users) Count() int {
	n := 0
	us.Range(func(u *User) {
		n = n + 1
	})

	return n
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
