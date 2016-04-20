package goio

import "sync"

func NewRooms() *rooms {
	rs := new(rooms)
	rs.rooms = make(map[string]*Room)
	rs.addRoom = make(chan *Room)
	rs.delRoom = make(chan *Room)

	rs.getRoom = make(chan string)
	rs.room = make(chan *Room)

	rs.getCount = make(chan struct{})
	rs.count = make(chan int)

	go func(rs *rooms) {
		for {
			select {
			case r := <-rs.addRoom:
				rs.rooms[r.Id] = r

			case r := <-rs.delRoom:
				r.closed = true
				rs.deleteRoom(r.Id)

			case roomId := <-rs.getRoom:
				room, _ := rs.rooms[roomId]
				if room != nil && room.closed {
					rs.deleteRoom(roomId)
					rs.room <- nil
				} else {
					rs.room <- room
				}

			case <-rs.getCount:
				rs.count <- len(rs.rooms)
			}
		}

	}(rs)

	return rs
}

type rooms struct {
	rooms   map[string]*Room
	addRoom chan *Room
	delRoom chan *Room
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

func (r *rooms) deleteRoom(roomId string) {
	delete(r.rooms, roomId)
}

func (r *rooms) Get(roomId string) *Room {
	r.getRoom <- roomId
	return <-r.room
}
