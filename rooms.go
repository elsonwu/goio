package goreal

type Rooms map[string]*Room

func (self *Rooms) Count() int {
	return len(*self)
}

func (self *Rooms) Delete(id string) {
	delete(*self, id)
}

func (self *Rooms) Add(room *Room) {
	if self.Has(room.Id) {
		return
	}

	(*self)[room.Id] = room
	room.On("broadcast", func(message *Message) {
		room.Receive(message)
	})

	room.On("destory", func(message *Message) {
		self.Delete(room.Id)
	})

	room.On("join", func(message *Message) {
		room.Receive(message)
	})

	room.On("leave", func(message *Message) {
		room.Receive(message)
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

	room := newRoom(id)
	GlobalRooms().Add(room)
	return room
}
