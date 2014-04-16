package goreal

import (
	"github.com/elsonwu/random"
	"log"
	"time"
)

var globalClients *Clients
var globalRooms *Rooms
var globalUsers *Users

func GlobalClients() *Clients {
	if globalClients == nil {
		globalClients = NewClients()
	}

	return globalClients
}

func GlobalRooms() *Rooms {
	if globalRooms == nil {
		globalRooms = NewRooms()
	}

	return globalRooms
}

func GlobalUsers() *Users {
	if globalUsers == nil {
		globalUsers = NewUsers()
	}

	return globalUsers
}

func NewUser(id string) *User {
	return &User{
		Id:      id,
		Clients: NewClients(),
		Rooms:   NewRooms(),
	}
}

func newRoom(id string) *Room {
	return &Room{
		Id:    id,
		Users: NewUsers(),
	}
}

func NewUsers() *Users {
	return &Users{}
}

func NewClients() *Clients {
	return &Clients{}
}

func NewRooms() *Rooms {
	return &Rooms{}
}

func Uuid() string {
	return random.String(20)
}

func NewClient() (clt *Client, done chan bool) {
	done = make(chan bool)
	clt = &Client{
		Id:            Uuid(),
		Msg:           make(chan *Message),
		LastHandshake: time.Now().Unix(),
	}

	go func(id string, done chan bool) {
		<-done

		for {
			time.Sleep(3 * time.Second)
			clt := GlobalClients().Get(id)
			if clt == nil {
				log.Printf("client is nil, id: %s", id)
				break
			}

			log.Printf("client id:%s, t:%d\n", clt.Id, time.Now().Unix()-clt.LastHandshake)
			if 30 < time.Now().Unix()-clt.LastHandshake {
				log.Printf("client id:%s destory \n", clt.Id)
				clt.Destory()
			}
		}
	}(clt.Id, done)

	return clt, done
}
