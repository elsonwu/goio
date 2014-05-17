package goio

type MClients struct {
	Clients []*Clients
	current *Clients
	max     int
}

func (self *MClients) Init() {
	if self.current == nil || self.Clients == nil {
		self.Clients = NewClients()
	}

	if self.current == nil || self.max < self.current.Count() {
		self.current = &Clients{
			Map: make(map[string]*Client),
		}
		self.Clients = append(self.Clients, self.current)
	}
}

func (self *MClients) Get(id string) *Client {
	for _, cs := range self.Clients {
		if clt := cs.Get(id); clt != nil {
			return clt
		}
	}

	return nil
}

func (self *MClients) Count() int {
	c := 0
	for _, cs := range self.Clients {
		c += cs.Count()
	}
	return c
}

func (self *MClients) Delete(id string) {
	for _, cs := range self.Clients {
		if clt := cs.Get(id); clt != nil {
			cs.Delete(id)
			return
		}
	}
}

func (self *MClients) Add(clt *Client) {
	if nil == self.Get(clt.Id) {
		self.Init()
		self.current.Add(clt)
	}
}
