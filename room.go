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

func (self *ClientRoom) Add(clt *Client) {
	if self.Has(clt.Id) {
		return
	}

	clt.On("destory", func(message *Message) {
		delete(self.Clients, clt.Id)
	})

	self.Clients[clt.Id] = clt
}
