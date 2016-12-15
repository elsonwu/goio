package goio

import (
	"sync"
	"time"

	"github.com/golang/glog"

	"gopkg.in/mgo.v2/bson"
)

func NewClientId() string {
	return bson.NewObjectId().Hex()
}

func NewClient(user *User) *Client {
	clt := &Client{
		Id:        NewClientId(),
		User:      user,
		messages:  make([]*Message, 0, 10),
		Handshake: make(chan bool),
	}

	AddUserClt(user, clt)

	// listen for client message
	go func(clt *Client) {
		for {
			select {
			case <-clt.Handshake:
				glog.V(2).Infof("client " + clt.Id + " handshake")
				// skip this waiting

			case <-time.After(time.Duration(LifeCycle) * time.Second):
				DelUserClt(clt.User, clt)
				return
			}
		}
	}(clt)

	return clt
}

type Client struct {
	Id        string
	User      *User
	messages  []*Message
	Handshake chan bool
	lock      sync.RWMutex
}

func (c *Client) AddMessage(msg *Message) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Handshake <- true
	c.messages = append(c.messages, msg)
}

func (c *Client) ReadMessages() []*Message {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Handshake <- true
	msgs := c.messages
	c.messages = make([]*Message, 0, 10)
	return msgs
}
