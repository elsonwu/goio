package goio

import (
	"fmt"
	"sync"
)

func NewRoom(roomId string) *Room {
	room := &Room{
		Id:      roomId,
		users:   make(map[string]*User),
		AddUser: make(chan *User),
		DelUser: make(chan *User),
		Message: make(chan *Message),

		getUserIds: make(chan struct{}),
		userIds:    make(chan []string),
	}

	Rooms().addRoom <- room

	go func(room *Room) {
		for {
			select {
			case u := <-room.AddUser:
				fmt.Printf("room %s added user %s \n", room.Id, u.Id)
				room.users[u.Id] = u
				u.addRoom <- room

			case u := <-room.DelUser:
				fmt.Printf("---------> room deleting user\n")

				delete(room.users, u.Id)

				go func(u *User) {
					u.delRoom <- room
				}(u)

				fmt.Printf("room %s deleted user %s, still have %d users \n", room.Id, u.Id, len(room.users))

				if len(room.users) == 0 {
					Rooms().delRoom <- room
					room.users = nil

					fmt.Printf("room %s deleted, break its loop\n", room.Id)
					//stop this loop
					return
				}

			case msg := <-room.Message:
				fmt.Printf("room %s received message from user %s client %s \n", room.Id, msg.CallerId, msg.ClientId)

				for _, u := range room.users {
					fmt.Printf("msg sent to user %s - start \n", u.Id)
					u.message <- msg
					fmt.Printf("msg sent to user %s - end \n", u.Id)
				}

			case <-room.getUserIds:
				uids := make([]string, 0, len(room.users))
				for _, u := range room.users {
					uids = append(uids, u.Id)
				}

				room.userIds <- uids
			}
		}

	}(room)

	fmt.Printf("created new room %s\n", room.Id)

	return room
}

type Room struct {
	Id         string
	users      map[string]*User
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
