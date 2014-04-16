package goio

import (
	"time"
)

type Client struct {
	Event
	Id            string
	User          *User
	Msg           chan *Message
	LastHandshake int64
	LifeCycle     int64
}

func (self *Client) Receive(message *Message) {
	self.Msg <- message
}

func (self *Client) Destory() {
	self.Emit("destory", &Message{
		EventName: "destory",
		Data:      self.Id,
	})
}

func (self *Client) Handshake() {
	self.LastHandshake = time.Now().Unix()
}

func (self *Client) LifeRemain() int64 {
	remain := self.LifeCycle - (time.Now().Unix() - self.LastHandshake)
	if 0 >= remain {
		return 0
	}

	return remain
}

func (self *Client) IsLive() bool {
	return 0 < self.LifeRemain()
}
