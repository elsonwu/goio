package goio

import (
	"sync"
)

// only for 10s for read/wirte chan
const roomWait = 10

func NewRoom(roomId string) *Room {
	return &Room{
		Id:      roomId,
		message: make(chan *Message),
		died:    false,
	}
}

type Room struct {
	Id        string
	m         sync.Map
	message   chan *Message
	died      bool
	userCount int
}

func (r *Room) AddMessage(msg *Message) {
	if r.died {
		return
	}

	var u *User
	r.m.Range(func(k interface{}, v interface{}) bool {
		u = v.(*User)
		if u == nil {
			return true
		}

		if !u.died {
			u.AddMessage(msg)
		}

		return true
	})
}

func (r *Room) AddUser(u *User) {
	if r.died {
		return
	}

	r.userCount = r.userCount + 1
	r.m.Store(u.Id, u)
}

func (r *Room) DelUser(u *User) {
	if r.died {
		return
	}

	r.userCount = r.userCount - 1
	r.m.Delete(u.Id)

	if r.userCount <= 0 {
		r.died = true
		Rooms().DelRoom(r)
	}
}

func (r *Room) UserIds() []string {
	if r.died {
		return nil
	}

	var userIds []string
	r.m.Range(func(k interface{}, v interface{}) bool {
		userIds = append(userIds, v.(string))
		return true
	})

	return userIds
}
