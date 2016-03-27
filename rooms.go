package goio

import "sync"

func NewRooms() *rooms {
	rs := new(rooms)
	rs.Rooms = make(map[string]*Room)
	rs.AddRoom = make(chan *Room)
	rs.DelRoom = make(chan *Room)

	rs.getRoom = make(chan string)
	rs.room = make(chan *Room)

	rs.getCount = make(chan struct{})
	rs.count = make(chan int)

	go func(rs *rooms) {
		for {
			select {
			case r := <-rs.AddRoom:
				rs.Rooms[r.Id] = r

			case r := <-rs.DelRoom:
				delete(rs.Rooms, r.Id)

			case roomId := <-rs.getRoom:
				room, _ := rs.Rooms[roomId]
				rs.room <- room

			case <-rs.getCount:
				rs.count <- len(rs.Rooms)
			}
		}

	}(rs)

	return rs
}

type rooms struct {
	Rooms   map[string]*Room
	AddRoom chan *Room
	DelRoom chan *Room
	room    chan *Room
	getRoom chan string

	count    chan int
	getCount chan struct{}
	lock     sync.RWMutex
}

func (r *rooms) Count() int {
	r.getCount <- struct{}{}
	return <-r.count
}

func (r *rooms) Get(roomId string) *Room {
	r.getRoom <- roomId
	return <-r.room
}
