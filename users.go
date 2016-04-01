package goio

import "sync"

func NewUsers() *users {
	us := new(users)
	us.Users = make(map[string]*User)
	us.addUser = make(chan *User)
	us.delUser = make(chan *User)

	us.getUser = make(chan string)
	us.user = make(chan *User)

	us.getCount = make(chan struct{})
	us.count = make(chan int)

	go func(us *users) {
		for {
			select {
			case u := <-us.addUser:
				us.Users[u.Id] = u

				// tell everyone this new client online
				go func(u *User) {
					Clients().Message <- &Message{
						EventName: "join",
						CallerId:  u.Id,
					}
				}(u)

			case u := <-us.delUser:
				delete(us.Users, u.Id)

				// tell everyone this user is offline
				go func(u *User) {
					Clients().Message <- &Message{
						EventName: "leave",
						CallerId:  u.Id,
					}
				}(u)

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
	addUser chan *User
	delUser chan *User

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
