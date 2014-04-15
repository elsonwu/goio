package goreal

type Rooms map[string]*Room

func (self *Rooms) Delete(id string) {
	delete(*self, id)
}

func (self *Rooms) Add(room *Room) {
	if self.Has(room.Id) {
		return
	}

	(*self)[room.Id] = room

	room.On("broadcast", func(message *Message) {
		for _, user := range room.Users {
			go user.Receive(message)
		}
	})

	room.On("destory", func(message *Message) {
		self.Delete(room.Id)
	})
}

func (self *Rooms) Has(id string) bool {
	_, ok := (*self)[id]
	return ok
}

func (self *Rooms) Get(id string) *Room {
	if room, ok := (*self)[id]; ok {
		return room
	}

	return NewRoom(id)
}
