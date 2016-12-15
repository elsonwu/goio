package goio

import (
	"time"

	"github.com/golang/glog"
)

func NewRoom(roomId string) *Room {
	room := &Room{
		Id:      roomId,
		users:   make(map[string]*User),
		addUser: make(chan *User),
		delUser: make(chan *User),
		Message: make(chan *Message),

		getUserIds: make(chan chan []string),
		close:      make(chan struct{}),
	}

	Rooms().addRoom <- room

	go func(room *Room) {
		for {
			select {
			case u := <-room.addUser:
				glog.V(3).Infof("room %s added user %s\n", room.Id, u.Id)
				room.users[u.Id] = u

			case u := <-room.delUser:
				glog.V(3).Infof("room %s deleting user %s\n", room.Id, u.Id)
				delete(room.users, u.Id)

				glog.V(3).Infof("room %s deleted user %s, still have %d users \n", room.Id, u.Id, len(room.users))
				if len(room.users) == 0 {
					Rooms().delRoom <- room

					go func(room *Room) {
						// wait 1s to receive all message
						time.Sleep(1 * time.Second)
						close(room.close)
					}(room)
				}

			case msg := <-room.Message:
				glog.V(3).Infof("room %s received message from user %s client %s \n", room.Id, msg.CallerId, msg.ClientId)
				for _, u := range room.users {
					u.message <- msg
					glog.V(3).Infof("msg sent to user %s\n", u.Id)
				}

			case userIds := <-room.getUserIds:
				uids := make([]string, 0, len(room.users))
				for _, u := range room.users {
					uids = append(uids, u.Id)
				}

				userIds <- uids

			case <-room.close:

				glog.V(3).Infof("room %s deleted, break its loop\n", room.Id)

				//stop this loop
				return

			}
		}

	}(room)

	glog.V(2).Infof("created new room %s\n", room.Id)

	return room
}

type Room struct {
	Id         string
	users      map[string]*User
	addUser    chan *User
	delUser    chan *User
	getUserIds chan chan []string
	Message    chan *Message
	close      chan struct{}
}

func (r *Room) UserIds() []string {
	userIds := make(chan []string)
	r.getUserIds <- userIds
	defer close(userIds)
	return <-userIds
}
