package goio

import (
	"sync"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func NewClientId() string {
	return bson.NewObjectId().Hex()
}

func NewClient(user *User) *Client {
	clt := &Client{
		Id:            NewClientId(),
		User:          user,
		Message:       make(chan *Message),
		LastHandshake: time.Now().Unix(),
	}

	go func() {
		for {
			select {
			case msg := <-clt.Message:
				// send message to the same user but other client
				clt.User.Message <- msg

			case <-time.After(10 * time.Second):
				// timeout
				close(clt.Message)

				clt.User.DelClt <- clt
				Clients().DelClt <- clt
				// stop this loop
				return
			}
		}
	}()

	return clt
}

type Client struct {
	Id            string
	User          *User
	Message       chan *Message
	LastHandshake int64
	LifeCycle     int64
	lock          sync.RWMutex
}

func (self *Client) Handshake() {
	self.LastHandshake = time.Now().Unix()
}
