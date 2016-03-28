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
		dataMap: make(map[string]string),
		data:    make(chan string),
		AddData: make(chan UserData),
		getData: make(chan string),
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
					close(user.AddData)
					close(user.data)
					close(user.getData)

					Users().DelUser <- user

					for _, r := range user.Rooms {
						r.DelUser <- user
					}

					user.Clients = nil
					user.Rooms = nil
					user.dataMap = nil

					// break this loop
					return
				}

			case room := <-user.AddRoom:
				user.Rooms[room.Id] = room

			case key := <-user.getData:
				user.data <- user.dataMap[key]

			case userData := <-user.AddData:
				user.dataMap[userData.Key] = userData.Val

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
	AddData chan UserData
	data    chan string
	getData chan string
	dataMap map[string]string
}

func (u *User) GetData(key string) string {
	u.getData <- key
	return <-u.data
}
