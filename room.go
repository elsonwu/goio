package goio

import (
	"time"

	"github.com/golang/glog"
)

// only for 10s for read/wirte chan
const roomWait = 10

func NewRoom(roomId string) *Room {
	room := &Room{
		Id:         roomId,
		users:      make(map[string]*User),
		addUser:    make(chan *User),
		delUser:    make(chan *User),
		message:    make(chan *Message),
		getUserIds: make(chan chan []string),
		died:       false,
	}

	Rooms().AddRoom(room)

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
					Rooms().DelRoom(room)
					room.died = true
					return
				}

			case msg := <-room.message:
				if room.died {
					continue
				}

				glog.V(3).Infof("room %s received message from user %s client %s \n", room.Id, msg.CallerId, msg.ClientId)
				for _, u := range room.users {
					if !u.died {
						u.AddMessage(msg)
						glog.V(3).Infof("msg sent to user %s\n", u.Id)
					}
				}

			case userIds := <-room.getUserIds:
				uids := make([]string, 0, len(room.users))
				for _, u := range room.users {
					uids = append(uids, u.Id)
				}

				userIds <- uids
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
	message    chan *Message
	died       bool
}

func (r *Room) AddMessage(msg *Message) {
	if r.died {
		return
	}

	select {
	case r.message <- msg:
	case <-time.After(roomWait * time.Second):
	}
}

func (r *Room) AddUser(u *User) {
	if r.died {
		return
	}

	select {
	case r.addUser <- u:
	case <-time.After(roomWait * time.Second):
	}
}

func (r *Room) DelUser(u *User) {
	if r.died {
		return
	}

	select {
	case r.delUser <- u:
	case <-time.After(roomWait * time.Second):
	}
}

func (r *Room) UserIds() []string {
	if r.died {
		return nil
	}

	userIds := make(chan []string)
	r.getUserIds <- userIds
	defer close(userIds)
	return <-userIds
}
