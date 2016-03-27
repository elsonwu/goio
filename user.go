package goio

func NewUser(userId string) *User {
	user := &User{
		Id:      userId,
		Clients: make(map[string]*Client),
		Rooms:   make(map[string]*Room),
		Message: make(chan *Message),
		AddClt:  make(chan *Client),
		DelClt:  make(chan *Client),
		AddRoom: make(chan *Room),
	}

	go func(user *User) {
		for {
			select {
			case msg := <-user.Message:
				for _, c := range user.Clients {
					if c.Id == msg.ClientId {
						continue
					}

					go func(c *Client, msg *Message) {
						c.Message <- msg
					}(c, msg)
				}

			case clt := <-user.AddClt:
				user.Clients[clt.Id] = clt

			case clt := <-user.DelClt:
				delete(user.Clients, clt.Id)

				if len(user.Clients) == 0 {
					close(user.AddClt)
					close(user.DelClt)
					close(user.Message)
					close(user.AddRoom)

					Users().DelUser <- user

					for _, r := range user.Rooms {
						r.DelUser <- user
					}

					user.Rooms = nil

					// break this loop
					return
				}

			case room := <-user.AddRoom:
				user.Rooms[room.Id] = room
			}
		}

	}(user)

	return user
}

type User struct {
	Id      string
	Clients map[string]*Client
	Rooms   map[string]*Room
	Message chan *Message
	AddClt  chan *Client
	DelClt  chan *Client
	AddRoom chan *Room
}
