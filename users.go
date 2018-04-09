package goio

import "sync"

func NewUsers() *users {
	us := new(users)
	return us
}

type users struct {
	m     sync.Map
	count int
}

func (us *users) AddUser(u *User) {
	if u.died {
		return
	}

	us.count = us.count + 1
	us.m.Store(u.Id, u)

	// tell everyone this new client online
	go func(userId string) {
		Clients().AddMessage(&Message{
			EventName: "join",
			CallerId:  userId,
		})
	}(u.Id)
}

func (us *users) DelUser(u *User) {
	us.count = us.count - 1
	us.m.Delete(u.Id)

	// tell everyone this user is offline
	go func(userId string) {
		Clients().AddMessage(&Message{
			EventName: "leave",
			CallerId:  userId,
		})
	}(u.Id)
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
		u.Run()
		Users().AddUser(u)
	}

	return u
}
