package goio

import (
	"time"

	"github.com/golang/glog"
)

func NewUser(userId string) *User {
	user := &User{
		Id:       userId,
		Clients:  make(map[string]*Client),
		rooms:    make(map[string]*Room),
		message:  make(chan *Message),
		addClt:   make(chan *Client),
		delClt:   make(chan *Client),
		addRoom:  make(chan *Room),
		delRoom:  make(chan *Room),
		dataMap:  make(map[string]string),
		data:     make(chan string),
		getRooms: make(chan chan map[string]*Room),
		AddData:  make(chan UserData),
		getData:  make(chan string),
		died:     false,
	}

	Users().AddUser(user)

	go func(user *User) {
		for {
			select {
			case msg := <-user.message:
				if user.died {
					continue
				}

				glog.V(3).Infof("user %s has %d clients, received message from user %s client %s\n", user.Id, len(user.Clients), msg.CallerId, msg.ClientId)
				for _, c := range user.Clients {
					if c.Id == msg.ClientId {
						glog.V(3).Infof("ignore message send from myself user[%s] client %s\n", c.User.Id, c.Id)
						continue
					}

					go c.AddMessage(msg)
					glog.V(3).Infof("received message to user[%s] client %s\n", c.User.Id, c.Id)
				}

			case clt := <-user.addClt:
				glog.V(3).Infof("user %s case addClt %s\n", user.Id, clt.Id)
				user.Clients[clt.Id] = clt

			case clt := <-user.delClt:
				delete(user.Clients, clt.Id)

				if len(user.Clients) == 0 {
					glog.V(3).Infof("## user %s has 0 client, need to del\n", user.Id)
					for _, r := range user.rooms {
						r.DelUser(user)
					}

					Users().DelUser(user)
					user.died = true
					return
				}
			case room := <-user.addRoom:
				glog.V(3).Infof("user case addRoom")
				user.rooms[room.Id] = room

			case room := <-user.delRoom:
				glog.V(3).Infof("user case delRoom")
				delete(user.rooms, room.Id)

			case key := <-user.getData:
				glog.V(3).Infof("user case getData")
				user.data <- user.dataMap[key]

			case userData := <-user.AddData:
				glog.V(3).Info("user case AddData")
				user.dataMap[userData.Key] = userData.Val

			case rooms := <-user.getRooms:
				rooms <- user.rooms
			}
		}

	}(user)

	return user
}

type User struct {
	Id       string
	Clients  map[string]*Client
	rooms    map[string]*Room
	message  chan *Message
	addClt   chan *Client
	delClt   chan *Client
	addRoom  chan *Room
	delRoom  chan *Room
	AddData  chan UserData
	getRooms chan chan map[string]*Room
	data     chan string
	getData  chan string
	dataMap  map[string]string
	died     bool
}

func (u *User) AddMessage(msg *Message) {
	if u.died {
		return
	}

	u.message <- msg
}

func (u *User) GetData(key string) string {
	if u.died {
		return ""
	}

	select {
	case u.getData <- key:
		return <-u.data
	case <-time.After(time.Second):
		return ""
	}
}

func (u *User) AddClt(clt *Client) {
	if u.died || clt.died {
		return
	}

	u.addClt <- clt
}

func (u *User) DelClt(clt *Client) {
	if u.died {
		return
	}

	u.delClt <- clt
}

func (u *User) AddRoom(room *Room) {
	if u.died || room.died {
		return
	}

	u.addRoom <- room
}

func (u *User) DelRoom(room *Room) {
	if u.died {
		return
	}

	u.delRoom <- room
}

func (u *User) Rooms() map[string]*Room {
	if u.died {
		return make(map[string]*Room)
	}

	rooms := make(chan map[string]*Room)
	u.getRooms <- rooms
	defer close(rooms)
	return <-rooms
}
