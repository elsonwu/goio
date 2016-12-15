package goio

func NewRooms() *rooms {
	rs := new(rooms)
	rs.rooms = make(map[string]*Room)
	rs.addRoom = make(chan *Room)
	rs.delRoom = make(chan *Room)
	rs.getRoom = make(chan roomGetter)
	rs.getCount = make(chan chan int)

	go func(rs *rooms) {
		for {
			select {
			case r := <-rs.addRoom:
				rs.rooms[r.Id] = r

			case r := <-rs.delRoom:
				delete(rs.rooms, r.Id)

			case roomGetter := <-rs.getRoom:
				r, _ := rs.rooms[roomGetter.roomId]
				roomGetter.room <- r

			case counter := <-rs.getCount:
				counter <- len(rs.rooms)
			}
		}

	}(rs)

	return rs
}

type roomGetter struct {
	roomId string
	room   chan *Room
}

type rooms struct {
	rooms    map[string]*Room
	addRoom  chan *Room
	delRoom  chan *Room
	getRoom  chan roomGetter
	getCount chan chan int
}

func (r *rooms) AddRoom(room *Room) {
	if room.died {
		return
	}

	r.addRoom <- room
}

func (r *rooms) DelRoom(room *Room) {
	r.delRoom <- room
}

func (r *rooms) Count() int {
	counter := make(chan int)
	r.getCount <- counter
	defer close(counter)
	return <-counter
}

func (r *rooms) Get(roomId string) *Room {
	room := make(chan *Room)
	r.getRoom <- roomGetter{
		roomId: roomId,
		room:   room,
	}
	defer close(room)
	return <-room
}

func (r *rooms) MustGet(roomId string) *Room {
	room := r.Get(roomId)
	if room == nil {
		room = NewRoom(roomId)
	}

	return room
}
