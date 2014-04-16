package goreal

type Room struct {
	Event
	Id    string
	Users *Users
}

func (self *Room) Has(id string) bool {
	return nil != self.Users.Get(id)
}

func (self *Room) Receive(message *Message) {
	for _, user := range *self.Users {
		go user.Receive(message)
	}
}

func (self *Room) Delete(id string) {
	delete(*self.Users, id)
	if 0 == self.Users.Count() {
		self.Destory()
	}

	self.Emit("broadcast", &Message{
		EventName: "leave",
		RoomId:    self.Id,
		CallerId:  id,
	})
}

func (self *Room) Destory() {
	self.Emit("destory", &Message{
		EventName: "destory",
		RoomId:    self.Id,
		CallerId:  self.Id,
	})
}

func (self *Room) Add(user *User) {
	if self.Has(user.Id) {
		return
	}

	user.Rooms.Add(self)
	user.On("destory", func(message *Message) {
		self.Delete(user.Id)
	})

	self.Users.Add(user)
}
