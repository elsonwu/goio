package goio

type MRooms struct {
	rooms   []*Rooms
	current *Rooms
}

func (self *MRooms) Init() {
	if self.current == nil || self.rooms == nil {
		self.rooms = NewRooms()
	}

	if self.current == nil || 100 < self.current.Count() {
		self.current = &Rooms{
			Map: make(map[string]*Room),
		}
		self.rooms = append(self.rooms, self.current)
	}
}

func (self *MRooms) Each(callback func(r *Room)) {
	for _, cs := range self.rooms {
		cs.Each(callback)
	}
}

func (self *MRooms) Get(id string, autoNew bool) *Room {
	for _, cs := range self.rooms {
		if clt := cs.Get(id, false); clt != nil {
			return clt
		}
	}

	if autoNew {
		self.Init()
		return self.current.Get(id, autoNew)
	}

	return nil
}

func (self *MRooms) Count() int {
	c := 0
	for _, cs := range self.rooms {
		c += cs.Count()
	}
	return c
}

func (self *MRooms) Delete(id string) {
	for _, cs := range self.rooms {
		if cs.Has(id) {
			cs.Delete(id)
			return
		}
	}
}

func (self *MRooms) Add(clt *Room) {
	if nil == self.Get(clt.Id, false) {
		self.Init()
		self.current.Add(clt)
	}
}
