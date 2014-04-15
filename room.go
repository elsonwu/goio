package goreal

type Room struct {
	Event
	Id    string
	Users Users
}

func (self *Room) Has(id string) bool {
	_, ok := self.Users[id]
	return ok
}

func (self *Room) Delete(id string) {
	delete(self.Users, id)
	if 0 == len(self.Users) {
		self.Destory()
	}

	self.Emit("broadcast", &Message{
		EventName: "leave",
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
