package goio

import "sync"

func NewUsers() *users {
	us := new(users)
	us.Users = make(map[string]*User)
	us.AddUser = make(chan *User)
	us.DelUser = make(chan *User)

	us.getUser = make(chan string)
	us.user = make(chan *User)

	us.getCount = make(chan struct{})
	us.count = make(chan int)

	go func(us *users) {
		for {
			select {
			case u := <-us.AddUser:
				us.Users[u.Id] = u

			case u := <-us.DelUser:
				delete(us.Users, u.Id)

			case userId := <-us.getUser:
				user, _ := us.Users[userId]
				us.user <- user

			case <-us.getCount:
				us.count <- len(us.Users)
			}
		}

	}(us)

	return us
}

type users struct {
	Users   map[string]*User
	AddUser chan *User
	DelUser chan *User

	getUser chan string
	user    chan *User // after GetUser

	getCount chan struct{}
	count    chan int
	lock     sync.RWMutex
}

func (r *users) Count() int {
	r.getCount <- struct{}{}
	return <-r.count
}

func (r *users) Get(userId string) *User {
	r.getUser <- userId
	return <-r.user
}

func (r *users) MustGet(userId string) *User {
	u := r.Get(userId)
	if u == nil {
		u = NewUser(userId)
	}

	return u
}
