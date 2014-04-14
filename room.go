package goreal

import ()

type ClientRoom struct {
	Event
	Id      string
	Clients map[string]*Client
}

func (self *ClientRoom) Has(id string) bool {
	_, ok := self.Clients[id]
	return ok
}

func (self *ClientRoom) Delete(id string) {
	delete(self.Clients, id)
	if 0 == len(self.Clients) {
		self.Destory()
	}

	self.Emit("broadcast", &Message{
		EventName: "left",
		Data:      id,
	})
}

func (self *ClientRoom) Destory() {
	self.Emit("destory", &Message{
		EventName: "destory",
		Data:      self.Id,
	})
}

func (self *ClientRoom) Add(clt *Client) {
	if self.Has(clt.Id) {
		return
	}

	clt.On("destory", func(message *Message) {
		self.Delete(clt.Id)
	})

	self.Clients[clt.Id] = clt
}
