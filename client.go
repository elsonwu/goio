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
		handshake: make(chan bool),
		died:      false,
	}

	AddUserClt(user, clt)

	// listen for client message
	go func(clt *Client) {
		for {
			select {
			case <-clt.handshake:
				glog.V(2).Infof("client " + clt.Id + " handshake")
				// skip this waiting

			case <-time.After(time.Duration(LifeCycle) * time.Second):
				DelUserClt(clt.User, clt)
				clt.lock.Lock()
				clt.died = true
				clt.lock.Unlock()
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
	handshake chan bool
	lock      sync.RWMutex
	died      bool
}

func (c *Client) AddMessage(msg *Message) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.died {
		return
	}

	c.messages = append(c.messages, msg)
}

func (c *Client) ReadMessages() []*Message {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.handshake <- true
	msgs := c.messages
	c.messages = make([]*Message, 0, 10)
	return msgs
}
