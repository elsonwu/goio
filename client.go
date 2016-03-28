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
		lastHandshake: time.Now().Unix(),
	}

	go func(clt *Client) {
		for {
			select {
			case msg := <-clt.Message:
				// send message to the same user but other client
				clt.User.Message <- msg

			case <-time.After(time.Duration(LifeCycle) * time.Second):

				// timeout, kill this client
				if time.Now().Unix()-clt.lastHandshake >= LifeCycle {
					close(clt.Message)

					clt.User.DelClt <- clt
					Clients().DelClt <- clt
					// stop this loop
					return
				}
			}
		}
	}(clt)

	return clt
}

type Client struct {
	Id            string
	User          *User
	Message       chan *Message
	lastHandshake int64
	lock          sync.RWMutex
}

func (self *Client) Handshake() {
	self.lastHandshake = time.Now().Unix()
}
