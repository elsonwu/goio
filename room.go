package goio

import "sync"

func NewRoom(roomId string) *Room {
	room := &Room{
		Id:      roomId,
		Users:   make(map[string]*User),
		AddUser: make(chan *User),
		DelUser: make(chan *User),
		Message: make(chan *Message),

		getUserIds: make(chan struct{}),
		userIds:    make(chan []string),
	}

	go func(room *Room) {
		for {
			select {
			case u := <-room.AddUser:
				room.Users[u.Id] = u

				u.AddRoom <- room

			case u := <-room.DelUser:
				delete(room.Users, u.Id)

				if len(room.Users) == 0 {
					close(room.AddUser)
					close(room.DelUser)
					close(room.Message)
					close(room.getUserIds)
					close(room.userIds)
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

			case <-room.getUserIds:
				uids := make([]string, 0, len(room.Users))
				for _, u := range room.Users {
					uids = append(uids, u.Id)
				}

				room.userIds <- uids
			}
		}

	}(room)

	return room
}

type Room struct {
	Id         string
	Users      map[string]*User
	AddUser    chan *User
	DelUser    chan *User
	getUserIds chan struct{}
	userIds    chan []string
	Message    chan *Message
	lock       sync.RWMutex
}

func (r *Room) UserIds() []string {
	r.getUserIds <- struct{}{}
	return <-r.userIds
}
