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
				go func(userId string) {
					Clients().Message <- &Message{
						EventName: "join",
						CallerId:  userId,
					}
				}(u.Id)

			case u := <-us.delUser:
				delete(us.Users, u.Id)

				// tell everyone this user is offline
				go func(userId string) {
					Clients().Message <- &Message{
						EventName: "leave",
						CallerId:  userId,
					}
				}(u.Id)

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

func (us *users) AddUser(u *User) {
	if u.died {
		return
	}

	us.addUser <- u
}

func (us *users) DelUser(u *User) {
	us.delUser <- u
}

func (us *users) Count() int {
	counter := make(chan int)
	us.getCount <- counter
	defer close(counter)
	return <-counter
}

func (us *users) Get(userId string) *User {
	user := make(chan *User)
	us.getUser <- userGetter{
		userId: userId,
		user:   user,
	}
	defer close(user)
	return <-user
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
