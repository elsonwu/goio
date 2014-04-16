package goreal

import (
	"github.com/elsonwu/random"
	"log"
	"time"
)

var UuidLen int = 20
var LifeCycle int64 = 60
var Debug bool = false
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
	return random.String(UuidLen)
}

func NewClient() (clt *Client, done chan bool) {
	done = make(chan bool)
	clt = &Client{
		Id:            Uuid(),
		Msg:           make(chan *Message),
		LastHandshake: time.Now().Unix(),
		LifeCycle:     LifeCycle,
	}

	go func(id string, done chan bool) {
		<-done

		for {
			time.Sleep(time.Duration(LifeCycle) * time.Second)
			clt := GlobalClients().Get(id)
			if clt == nil {
				if Debug {
					log.Printf("client is nil, id: %s", id)
				}

				break
			}

			if Debug {
				log.Printf("client id:%s, t:%d\n", clt.Id, clt.LifeRemain())
			}

			if !clt.IsLive() {
				if Debug {
					log.Printf("client id:%s destory \n", clt.Id)
				}

				clt.Destory()
			}
		}
	}(clt.Id, done)

	return clt, done
}
