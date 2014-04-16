package goio

type callback func(message *Message)

type Event struct {
	evts map[string][]callback
}

func (self *Event) On(eventName string, fn callback) {
	if self.evts == nil {
		self.evts = make(map[string][]callback)
	}

	if _, ok := self.evts[eventName]; !ok {
		self.evts[eventName] = make([]callback, 0)
	}

	self.evts[eventName] = append(self.evts[eventName], fn)
}

func (self *Event) Emit(eventName string, message *Message) {
	if _, ok := self.evts[eventName]; ok {
		for _, fn := range self.evts[eventName] {
			fn(message)
		}
	}
}
