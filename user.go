package goio

import (
	"sync"

	"github.com/golang/glog"
)

func newUser(userId string) *User {
	user := &User{
		Id: userId,
	}

	Users().addUser(user)

	SendMessage(&Message{
		EventName: MsgJoin,
		CallerId:  user.Id,
	}, user)

	return user
}

type User struct {
	Id string

	clients      sync.Map
	clientsCount int

	rooms      sync.Map
	roomsCount int

	data sync.Map
}

func (u *User) IsDead() bool {
	if u.clientsCount >= 2 {
		return false
	}

	return !u.anyActiveClient()
}

func (u *User) addMessage(msg *Message) {
	if u.IsDead() {
		return
	}

	glog.V(1).Infoln("user.addMessage " + u.Id + " message " + msg.EventName)
	u.clients.Range(func(k interface{}, v interface{}) bool {
		c := v.(*Client)
		if c == nil || c.IsDead() {
			return true
		}

		if c.Id == msg.ClientId {
			return true
		}

		c.addMessage(msg)
		return true
	})
}

func (u *User) AddData(k string, v string) {
	u.data.Store(k, v)
}

func (u *User) GetData(key string) string {
	if u.IsDead() {
		return ""
	}

	v, ok := u.data.Load(key)
	if !ok {
		return ""
	}

	return v.(string)
}

func (u *User) ClientCount() int {
	return u.clientsCount
}

func (u *User) anyActiveClient() bool {
	has := false
	u.clients.Range(func(k interface{}, v interface{}) bool {
		c, ok := v.(*Client)
		if c == nil || !ok {
			return true
		}

		if c.IsDead() {
			has = true
			return false
		}

		return true
	})

	return has
}

func (u *User) AddClt(clt *Client) {
	u.clientsCount += 1
	u.clients.Store(clt.Id, clt)
}

func (u *User) DelClt(clt *Client) {
	u.clientsCount -= 1
	u.clients.Delete(clt.Id)
}

func (u *User) AddRoom(room *Room) {
	if u.IsDead() {
		return
	}

	u.roomsCount += 1
	u.rooms.Store(room.Id, room)
}

func (u *User) DelRoom(room *Room) {
	u.roomsCount -= 1
	u.rooms.Delete(room.Id)
}

func (u *User) Rooms() map[string]*Room {
	if u.roomsCount <= 0 {
		return nil
	}

	rooms := make(map[string]*Room)
	u.rooms.Range(func(k interface{}, v interface{}) bool {
		room := v.(*Room)
		rooms[room.Id] = room
		return true
	})

	return rooms
}
