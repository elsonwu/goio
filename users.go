package goreal

type Users map[string]*User

func (self *Users) Add(user *User) {
	if nil != self.Get(user.Id) {
		return
	}

	user.On("destory", func(message *Message) {
		self.Delete(user.Id)
	})

	(*self)[user.Id] = user
}

func (self *Users) Delete(userId string) {
	delete(*self, userId)
}

func (self *Users) Get(userId string) *User {
	if user, ok := (*self)[userId]; ok {
		return user
	}

	return nil
}
