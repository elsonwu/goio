package goio

type MUsers struct {
	Users   []*Users
	current *Users
}

func (self *MUsers) Init() {
	if self.current == nil || self.Users == nil {
		self.Users = NewUsers()
	}

	if self.current == nil || 1000 < self.current.Count() {
		self.current = &Users{
			Map: make(map[string]*User),
		}
		self.Users = append(self.Users, self.current)
	}
}

func (self *MUsers) Get(id string) *User {
	for _, cs := range self.Users {
		if clt := cs.Get(id); clt != nil {
			return clt
		}
	}

	return nil
}

func (self *MUsers) Count() int {
	c := 0
	for _, cs := range self.Users {
		c += cs.Count()
	}
	return c
}

func (self *MUsers) Delete(id string) {
	for _, cs := range self.Users {
		if clt := cs.Get(id); clt != nil {
			cs.Delete(id)
			return
		}
	}
}

func (self *MUsers) Add(clt *User) {
	if nil == self.Get(clt.Id) {
		self.Init()
		self.current.Add(clt)
	}
}
