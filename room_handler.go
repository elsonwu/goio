package goreal

type ClientRoomHandler map[string]*ClientRoom

func (self *ClientRoomHandler) deleteRoom(id string) {
	delete(*self, id)
}

func (self *ClientRoomHandler) newRoom(id string) *ClientRoom {
	cr := &ClientRoom{
		Id:      id,
		Clients: make(map[string]*Client),
	}

	(*self)[cr.Id] = cr

	cr.On("broadcast", func(message *Message) {
		for _, clt := range cr.Clients {
			go clt.Receive(message)
		}
	})

	cr.On("destory", func(message *Message) {
		self.deleteRoom(cr.Id)
	})

	return cr
}

func (self *ClientRoomHandler) Room(id string) *ClientRoom {
	room, ok := (*self)[id]
	if ok {
		return room
	}

	return self.newRoom(id)
}
