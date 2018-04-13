package goio

import (
	"sync"
)

// only for 10s for read/wirte chan
const roomWait = 10

func NewRoom(roomId string) *Room {
	r := &Room{
		Id: roomId,
	}

	Rooms().AddRoom(r)

	return r
}

type Room struct {
	Id        string
	m         sync.Map
	userCount int
}

func (r *Room) IsDead() bool {
	if r.userCount >= 2 {
		return false
	}

	return !r.anyActiveUser()
}

func (r *Room) anyActiveUser() bool {
	has := false
	r.m.Range(func(k interface{}, v interface{}) bool {
		u, ok := v.(*User)
		if u == nil || !ok {
			return true
		}

		if u.anyActiveClient() {
			has = true
			return false
		}

		return true
	})

	return has
}

func (r *Room) UserCount() int {
	return r.userCount
}

func (r *Room) addMessage(msg *Message) {
	if r.IsDead() {
		return
	}

	r.m.Range(func(k interface{}, v interface{}) bool {
		u := v.(*User)
		if u == nil || u.IsDead() {
			return true
		}

		u.addMessage(msg)
		return true
	})
}

func (r *Room) addUser(u *User) {
	r.userCount += 1
	r.m.Store(u.Id, u)
}

func (r *Room) delUser(u *User) {
	r.userCount -= 1
	r.m.Delete(u.Id)
}

func (r *Room) UserIds() []string {
	if r.IsDead() {
		return nil
	}

	var userIds []string
	r.m.Range(func(k interface{}, v interface{}) bool {
		userIds = append(userIds, v.(*User).Id)
		return true
	})

	return userIds
}
