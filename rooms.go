package goio

import "sync"

func NewRooms() *rooms {
	rs := new(rooms)
	return rs
}

type rooms struct {
	m     sync.Map
	count int
}

func (r *rooms) AddRoom(room *Room) {
	if room.died {
		return
	}

	r.count = r.count + 1
	r.m.Store(room.Id, room)
}

func (r *rooms) DelRoom(room *Room) {
	r.count = r.count - 1
	r.m.Delete(room.Id)
}

func (r *rooms) Count() int {
	return r.count
}

func (r *rooms) Get(roomId string) *Room {
	v, ok := r.m.Load(roomId)
	if !ok {
		return nil
	}

	return v.(*Room)
}

func (r *rooms) MustGet(roomId string) *Room {
	room := r.Get(roomId)
	if room == nil {
		room = NewRoom(roomId)
		r.AddRoom(room)
	}

	return room
}
