package goio

import (
	"github.com/elsonwu/random"
	"log"
	"time"
)

var UuidLen int = 12
var LifeCycle int64 = 60
var Debug bool = false
var GClients IClients
var GRooms IRooms
var GUsers IUsers

func GlobalClients() IClients {
	if GClients == nil {
		GClients = NewMClients()
	}

	return GClients
}

func GlobalRooms() IRooms {
	if GRooms == nil {
		GRooms = NewMRooms()
	}

	return GRooms
}

func GlobalUsers() IUsers {
	if GUsers == nil {
		GUsers = NewMUsers()
	}

	return GUsers
}

func NewUser(id string) *User {
	user := &User{
		Id: id,
		ClientIds: MapBool{
			Map: make(mapBool),
		},
		RoomIds: MapBool{
			Map: make(mapBool),
		},
	}

	user.On("join", func(message *Message) {
		if user.RoomIds.Has(message.RoomId) {
			return
		}

		GlobalRooms().Get(message.RoomId, true).Add(user)
	})

	user.On("leave", func(message *Message) {
		if !user.RoomIds.Has(message.RoomId) {
			return
		}

		room := GlobalRooms().Get(message.RoomId, false)
		if room == nil {
			return
		}

		room.Delete(user.Id)
	})

	user.On("broadcast", func(message *Message) {
		if message.RoomId == "" {
			for roomId, _ := range user.RoomIds.Map {
				room := GlobalRooms().Get(roomId, true)
				room.Receive(message)
			}
		} else {
			room := GlobalRooms().Get(message.RoomId, true)
			room.Receive(message)
		}
	})

	GlobalUsers().Add(user)
	return user
}

func NewRoom(id string) *Room {
	room := &Room{
		Id: id,
		UserIds: MapBool{
			Map: make(mapBool),
		},
	}

	GlobalRooms().Add(room)
	return room
}

func NewUsers() []*Users {
	return make([]*Users, 0, 10)
}

func NewMUsers() *MUsers {
	return &MUsers{
		Users: NewUsers(),
		max:   1000,
	}
}

func NewClients() []*Clients {
	return make([]*Clients, 0, 10)
}

func NewMClients() *MClients {
	return &MClients{
		Clients: NewClients(),
		max:     1000,
	}
}

func NewRooms() []*Rooms {
	return make([]*Rooms, 0, 10)
}

func NewMRooms() *MRooms {
	return &MRooms{
		Rooms: NewRooms(),
		max:   100,
	}
}

func NewMessages() []*Message {
	return make([]*Message, 0, 20)
}

func Uuid() string {
	return random.String(UuidLen)
}

func NewClientId() string {
	uuid := Uuid()
	if GlobalClients().Get(uuid) != nil {
		return NewClientId()
	}

	return uuid
}

func NewClient() (clt *Client, done chan bool) {
	clt = &Client{
		Id:            NewClientId(),
		Messages:      NewMessages(),
		LastHandshake: time.Now().Unix(),
		LifeCycle:     LifeCycle,
	}

	GlobalClients().Add(clt)
	done = make(chan bool)
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
					log.Printf("client id:%s destroy \n", clt.Id)
				}

				clt.Destroy()
			}
		}
	}(clt.Id, done)

	return clt, done
}
