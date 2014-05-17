package goio

type IRooms interface {
	Each(callback func(*Room))
	Get(id string, autoNew bool) *Room
	Delete(id string)
	Add(*Room)
	Count() int
}

type IUsers interface {
	Each(callback func(*User))
	Get(id string) *User
	Count() int
	Delete(id string)
	Add(*User)
}

type IClients interface {
	Each(callback func(*Client))
	Get(id string) *Client
	Count() int
	Delete(id string)
	Add(*Client)
}
