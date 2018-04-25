package goio

import "sync"

func NewRooms() *rooms {
	rs := new(rooms)
	return rs
}

type rooms struct {
	m sync.Map
}

func (r *rooms) AddRoom(room *Room) {
	r.m.Store(room.Id, room)
}

func (r *rooms) DelRoom(roomId string) {
	r.m.Delete(roomId)
}

func (r *rooms) Count() int {
	n := 0
	r.Range(func(r *Room) {
		n = n + 1
	})

	return n
}

func (r *rooms) Range(f func(r *Room)) {
	r.m.Range(func(k interface{}, v interface{}) bool {
		room, ok := v.(*Room)
		if !ok {
			return true
		}

		f(room)
		return true
	})
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
	}

	return room
}
