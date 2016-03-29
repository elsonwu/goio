package goio

import "fmt"

func NewUser(userId string) *User {
	user := &User{
		Id:      userId,
		Clients: make(map[string]*Client),
		rooms:   make(map[string]*Room),
		message: make(chan *Message),
		addClt:  make(chan *Client),
		DelClt:  make(chan *Client),
		addRoom: make(chan *Room),
		delRoom: make(chan *Room),
		dataMap: make(map[string]string),
		data:    make(chan string),
		AddData: make(chan UserData),
		getData: make(chan string),
	}

	Users().AddUser <- user

	go func(user *User) {
		for {
			select {
			case msg, ok := <-user.message:
				if !ok {
					return
				}

				// fmt.Printf("user %s has %d clients, received message from user %s client %s\n", user.Id, len(user.Clients), msg.CallerId, msg.ClientId)
				for _, c := range user.Clients {
					if c.Id == msg.ClientId {
						fmt.Printf("ignore message send from myself user[%s] client %s\n", c.User.Id, c.Id)
						continue
					}

					// fmt.Printf("sending message to user[%s] client %s - start \n", c.User.Id, c.Id)
					c.message <- msg
					// fmt.Printf("sending message to user[%s] client %s - end \n", c.User.Id, c.Id)
				}

			case clt, ok := <-user.addClt:

				if !ok {
					return
				}

				fmt.Printf("#### user %s case addClt %s\n", user.Id, clt.Id)
				user.Clients[clt.Id] = clt

			case clt, ok := <-user.DelClt:

				if !ok {
					return
				}

				delete(user.Clients, clt.Id)
				fmt.Printf("user %s case delClt - %s, still has %d clients \n", user.Id, clt.Id, len(user.Clients))

				Clients().delClt <- clt

				if len(user.Clients) == 0 {
					fmt.Printf("## user %s has 0 client, need to del\n", user.Id)

					Users().DelUser <- user
					for _, r := range user.rooms {
						r.DelUser <- user
					}

					user.Clients = nil
					user.rooms = nil
					user.dataMap = nil

					fmt.Printf("-------> user %s deleted, break its loop\n", user.Id)
					return
				} else {
					fmt.Printf("## user %s has %d client\n", user.Id, len(user.Clients))
				}

			case room, ok := <-user.addRoom:

				if !ok {
					return
				}

				fmt.Println("user case addRoom")
				user.rooms[room.Id] = room

			case room, ok := <-user.delRoom:
				if !ok {
					return
				}

				fmt.Println("user case delRoom")
				delete(user.rooms, room.Id)

			case key, ok := <-user.getData:
				if !ok {
					return
				}

				fmt.Println("user case getData")
				user.data <- user.dataMap[key]

			case userData, ok := <-user.AddData:
				if !ok {
					return
				}
				fmt.Println("user case AddData")
				user.dataMap[userData.Key] = userData.Val

			case <-user.getRooms:
				user.rs <- user.rooms
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
	DelClt   chan *Client
	addRoom  chan *Room
	delRoom  chan *Room
	AddData  chan UserData
	getRooms chan struct{}
	rs       chan map[string]*Room
	data     chan string
	getData  chan string
	dataMap  map[string]string
}

func (u *User) GetData(key string) string {
	u.getData <- key
	return <-u.data
}

func (u *User) Rooms() map[string]*Room {
	u.getRooms <- struct{}{}
	return <-u.rs
}
