package goio

import (
	"fmt"
	"sync"
)

func NewRoom(roomId string) *Room {
	room := &Room{
		Id:      roomId,
		Users:   make(map[string]*User),
		AddUser: make(chan *User),
		DelUser: make(chan *User),
		Message: make(chan *Message),
	}

	go func(room *Room) {
		for {
			select {
			case u := <-room.AddUser:
				room.Users[u.Id] = u

				u.AddRoom <- room

			case u := <-room.DelUser:
				delete(room.Users, u.Id)

				fmt.Println("room del user")

				if len(room.Users) == 0 {
					close(room.AddUser)
					close(room.DelUser)
					close(room.Message)
					room.Users = nil

					Rooms().DelRoom <- room

					//stop this loop
					return
				}

			case msg := <-room.Message:
				for _, u := range room.Users {
					go func(u *User, msg *Message) {
						u.Message <- msg
					}(u, msg)
				}
			}
		}

	}(room)

	return room
}

type Room struct {
	Id      string
	Users   map[string]*User
	AddUser chan *User
	DelUser chan *User
	Message chan *Message
	lock    sync.RWMutex
}

func (r *Room) UserIds() []string {
	r.lock.RLock()
	defer r.lock.RUnlock()

	uids := make([]string, 0, len(r.Users))
	for _, u := range r.Users {
		uids = append(uids, u.Id)
	}

	return uids
}
