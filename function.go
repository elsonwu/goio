package goio

import (
	"github.com/elsonwu/random"
	"log"
	"time"
)

var UuidLen int = 20
var LifeCycle int64 = 60
var Debug bool = false
var GClients *Clients
var GRooms *Rooms
var GUsers *Users

func GlobalClients() *Clients {
	if GClients == nil {
		GClients = NewClients()
	}

	return GClients
}

func GlobalRooms() *Rooms {
	if GRooms == nil {
		GRooms = NewRooms()
	}

	return GRooms
}

func GlobalUsers() *Users {
	if GUsers == nil {
		GUsers = NewUsers()
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

func NewUsers() *Users {
	return &Users{
		Map: make(map[string]*User),
	}
}

func NewClients() *Clients {
	return &Clients{
		Map: make(map[string]*Client),
	}
}

func NewRooms() *Rooms {
	return &Rooms{
		Map: make(map[string]*Room),
	}
}

func Uuid() string {
	return random.String(UuidLen)
}

func NewClient() (clt *Client, done chan bool) {
	clt = &Client{
		Id:            Uuid(),
		Messages:      make([]*Message, 0, 20),
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
				log.Printf("client id:%s destroy \n", clt.Id)
				clt.Destroy()
			}
		}
	}(clt.Id, done)

	return clt, done
}
