package goio

type MRooms struct {
	Rooms   []*Rooms
	current *Rooms
}

func (self *MRooms) Init() {
	if self.current == nil || self.Rooms == nil {
		self.Rooms = NewRooms()
	}

	if self.current == nil || 100 < self.current.Count() {
		self.current = &Rooms{
			Map: make(map[string]*Room),
		}
		self.Rooms = append(self.Rooms, self.current)
	}
}

func (self *MRooms) Each(callback func(r *Room)) {
	for _, cs := range self.Rooms {
		cs.Each(callback)
	}
}

func (self *MRooms) Get(id string, autoNew bool) *Room {
	for _, cs := range self.Rooms {
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
	for _, cs := range self.Rooms {
		c += cs.Count()
	}
	return c
}

func (self *MRooms) Delete(id string) {
	for _, cs := range self.Rooms {
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
