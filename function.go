package goreal

import (
	"log"
	"strconv"
	"time"
)

var globalClients *Clients
var globalRooms *Rooms
var globalUsers *Users

func GlobalClients() *Clients {
	if globalClients == nil {
		globalClients = &Clients{}
	}

	return globalClients
}

func GlobalRooms() *Rooms {
	if globalRooms == nil {
		globalRooms = &Rooms{}
	}

	return globalRooms
}

func GlobalUsers() *Users {
	if globalUsers == nil {
		globalUsers = &Users{}
	}

	return globalUsers
}

func NewUser(id string) *User {
	return &User{
		Id:      id,
		Clients: make(map[string]*Client),
		Rooms:   make(Rooms),
	}
}

func NewRoom(id string) *Room {

	if GlobalRooms().Has(id) {
		return GlobalRooms().Get(id)
	}

	room := &Room{
		Id:    id,
		Users: make(Users),
	}

	GlobalRooms().Add(room)
	return room
}

func NewRooms() Rooms {
	return make(Rooms)
}

func NewClients() Clients {
	return make(Clients)
}

func Uuid() string {
	return strconv.Itoa(int(time.Now().Nanosecond()))
}

func NewClient() *Client {
	clt := &Client{
		Id:            Uuid(),
		Msg:           make(chan *Message),
		LastHandshake: time.Now().Unix(),
	}

	go func(id string) {
		for {
			clt := GlobalClients().Get(id)
			if clt == nil {
				break
			}

			time.Sleep(3 * time.Second)
			if 60 < time.Now().Unix()-clt.LastHandshake {
				log.Printf("Client id: %s destory \n", clt.Id)
				clt.Destory()
			}
		}
	}(clt.Id)

	return clt
}
