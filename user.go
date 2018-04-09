package goio

import (
	"sync"
	"time"
)

func newUser(userId string) *User {
	user := &User{
		Id:      userId,
		message: make(chan *Message),
		died:    false,
	}

	return user
}

type User struct {
	Id      string
	message chan *Message

	clients      sync.Map
	clientsCount int

	rooms      sync.Map
	roomsCount int

	data sync.Map
	died bool
}

func (u *User) Run() {
	go func(user *User) {
		for {
			select {
			case msg := <-user.message:
				if user.died {
					continue
				}

				user.clients.Range(func(k interface{}, v interface{}) bool {
					c := v.(*Client)
					if c.Id == msg.ClientId {
						return true
					}

					go c.AddMessage(msg)
					return true
				})
			}
		}

	}(u)
}

func (u *User) AddMessage(msg *Message) {
	if u.died {
		return
	}

	select {
	case u.message <- msg:
	case <-time.After(time.Second):
	}
}

func (u *User) AddData(k string, v string) {
	u.data.Store(k, v)
}

func (u *User) GetData(key string) string {
	if u.died {
		return ""
	}

	v, ok := u.data.Load(key)
	if !ok {
		return ""
	}

	return v.(string)
}

func (u *User) AddClt(clt *Client) {
	if u.died || clt.died {
		return
	}

	u.clientsCount = u.clientsCount + 1
	u.clients.Store(clt.Id, clt)
}

func (u *User) DelClt(clt *Client) {
	if u.died {
		return
	}

	u.clientsCount = u.clientsCount - 1
	u.clients.Delete(clt.Id)

	if u.clientsCount <= 0 {
		u.died = true
		Users().DelUser(u)
		u.rooms.Range(func(k interface{}, v interface{}) bool {
			v.(*Room).DelUser(u)
			return true
		})
	}
}

func (u *User) AddRoom(room *Room) {
	if u.died || room.died {
		return
	}

	u.roomsCount = u.roomsCount + 1
	u.rooms.Store(room.Id, room)
}

func (u *User) DelRoom(room *Room) {
	if u.died {
		return
	}

	u.roomsCount = u.roomsCount + 1
	u.rooms.Delete(room.Id)
}

func (u *User) Rooms() map[string]*Room {
	if u.died {
		return make(map[string]*Room)
	}

	var rooms map[string]*Room
	u.rooms.Range(func(k interface{}, v interface{}) bool {
		room := v.(*Room)
		rooms[room.Id] = room
		return true
	})

	return rooms
}
