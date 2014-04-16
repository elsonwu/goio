package goio

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
	user := &User{
		Id:      id,
		Clients: NewClients(),
		Rooms:   NewRooms(),
	}

	user.Emit("broadcast", &Message{
		EventName: "connect",
		CallerId:  user.Id,
	})

	user.On("destory", func(message *Message) {
		user.Emit("disconnect", &Message{
			EventName: "disconnect",
			CallerId:  user.Id,
		})
	})

	user.On("disconnect", func(message *Message) {
		for _, room := range *user.Rooms {
			room.Receive(message)
		}
	})

	user.On("join", func(message *Message) {
		if message.RoomId == "" {
			return
		}

		room := GlobalRooms().Get(message.RoomId)
		if !room.Has(user.Id) {
			room.Emit("broadcast", &Message{
				EventName: "join",
				RoomId:    room.Id,
				Data:      user.Id,
			})

			user.On("destory", func(message *Message) {
				room.Emit("broadcast", &Message{
					EventName: "leave",
					Data:      user.Id,
				})
			})

			room.Add(user)
		}
	})

	user.On("leave", func(message *Message) {
		if roomId, ok := message.Data.(string); ok {
			room := GlobalRooms().Get(roomId)
			if room.Has(user.Id) {
				room.Emit("broadcast", &Message{
					EventName: "leave",
					RoomId:    room.Id,
					Data:      user.Id,
				})
				room.Delete(user.Id)
			}
		}
	})

	user.On("broadcast", func(message *Message) {
		if message.RoomId == "" {
			for _, room := range *user.Rooms {
				room.Emit("broadcast", message)
			}
		} else {
			GlobalRooms().Get(message.RoomId).Emit("broadcast", message)
		}
	})

	return user
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
