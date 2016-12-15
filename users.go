package goio

func NewUsers() *users {
	us := new(users)
	us.Users = make(map[string]*User)
	us.addUser = make(chan *User)
	us.delUser = make(chan *User)

	us.getUser = make(chan userGetter)

	us.getCount = make(chan chan int)

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

			case userGetter := <-us.getUser:
				user, _ := us.Users[userGetter.userId]
				userGetter.user <- user

			case counter := <-us.getCount:
				counter <- len(us.Users)
			}
		}

	}(us)

	return us
}

type userGetter struct {
	userId string
	user   chan *User
}

type users struct {
	Users    map[string]*User
	addUser  chan *User
	delUser  chan *User
	getUser  chan userGetter
	getCount chan chan int
}

func (r *users) Count() int {
	counter := make(chan int)
	r.getCount <- counter
	defer close(counter)
	return <-counter
}

func (r *users) Get(userId string) *User {
	user := make(chan *User)
	r.getUser <- userGetter{
		userId: userId,
		user:   user,
	}
	defer close(user)
	return <-user
}

func (r *users) MustGet(userId string) *User {
	u := r.Get(userId)
	if u == nil {
		u = NewUser(userId)
	}

	return u
}
