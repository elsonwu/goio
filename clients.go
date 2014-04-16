package goreal

type Clients map[string]*Client

func (self *Clients) Get(id string) *Client {
	if clt, ok := (*self)[id]; ok {
		return clt
	}

	return nil
}

func (self *Clients) Count() int {
	return len(*self)
}

func (self *Clients) Receive(message *Message) {
	for _, clt := range *self {
		go func(clt *Client, msg *Message) {
			clt.Msg <- msg
		}(clt, message)
	}
}

func (self *Clients) Delete(id string) {
	delete(*self, id)
}

func (self *Clients) Add(clt *Client) {
	if nil != self.Get(clt.Id) {
		return
	}

	(*self)[clt.Id] = clt
	clt.On("destory", func(message *Message) {
		self.Delete(clt.Id)
	})
}
