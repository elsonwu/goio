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
		getRooms: make(chan struct{}),
		rs:       make(chan map[string]*Room),
		AddData:  make(chan UserData),
		getData:  make(chan string),
		close:    make(chan struct{}),
	}

	Users().addUser <- user

	go func(user *User) {
		for {
			select {
			case msg, ok := <-user.message:
				if !ok {
					return
				}

				glog.V(3).Infof("user %s has %d clients, received message from user %s client %s\n", user.Id, len(user.Clients), msg.CallerId, msg.ClientId)
				for _, c := range user.Clients {
					if c.Id == msg.ClientId {
						glog.V(3).Infof("ignore message send from myself user[%s] client %s\n", c.User.Id, c.Id)
						continue
					}

					c.receiveMessage <- msg
					glog.V(3).Infof("received message to user[%s] client %s\n", c.User.Id, c.Id)
				}

			case clt, ok := <-user.addClt:

				if !ok {
					return
				}

				glog.V(3).Infof("user %s case addClt %s\n", user.Id, clt.Id)
				user.Clients[clt.Id] = clt

			case clt, ok := <-user.delClt:

				if !ok {
					return
				}

				delete(user.Clients, clt.Id)

				if len(user.Clients) == 0 {

					glog.V(3).Infof("## user %s has 0 client, need to del\n", user.Id)
					for _, r := range user.rooms {
						r.delUser <- user
					}

					Users().delUser <- user

					go func() {
						// wait 1s to receive all messages
						time.Sleep(1 * time.Second)
						close(user.close)
					}()
				}
			case room, ok := <-user.addRoom:

				if !ok {
					return
				}

				glog.V(3).Infof("user case addRoom")
				user.rooms[room.Id] = room

			case room, ok := <-user.delRoom:
				if !ok {
					return
				}

				glog.V(3).Infof("user case delRoom")
				delete(user.rooms, room.Id)

			case key, ok := <-user.getData:
				if !ok {
					return
				}

				glog.V(3).Infof("user case getData")
				user.data <- user.dataMap[key]

			case userData, ok := <-user.AddData:
				if !ok {
					return
				}

				glog.V(3).Info("user case AddData")
				user.dataMap[userData.Key] = userData.Val

			case <-user.getRooms:
				user.rs <- user.rooms

			case <-user.close:
				// wait for GC
				// close(user.AddData)
				// close(user.addClt)
				// close(user.addRoom)
				// close(user.data)
				// close(user.delClt)
				// close(user.delRoom)
				// close(user.getData)
				// close(user.getRooms)
				// close(user.message)
				// close(user.rs)
				// user.Clients = nil
				// user.dataMap = nil
				// user.rooms = nil

				// break this loop
				glog.V(3).Infof("user %s deleted, break its loop\n", user.Id)
				return
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
	getRooms chan struct{}
	rs       chan map[string]*Room
	data     chan string
	getData  chan string
	dataMap  map[string]string
	close    chan struct{}
}

func (u *User) GetData(key string) string {
	select {
	case u.getData <- key:
		return <-u.data
	case <-time.After(time.Second):
		return ""
	}
}

func (u *User) Rooms() map[string]*Room {
	select {
	case u.getRooms <- struct{}{}:
		return <-u.rs
	case <-time.After(time.Second):
		return nil
	}
}
