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
		Id:             NewClientId(),
		User:           user,
		receiveMessage: make(chan *Message),
		messages:       make([]*Message, 0, 10),
		close:          make(chan struct{}),
		fetchMessages:  make(chan struct{}),
		msgs:           make(chan []*Message),
		lastHandshake:  time.Now().Unix(),
	}

	AddUserClt(user, clt)

	// check for client life cycle
	go func(clt *Client) {
		for {
			time.Sleep(time.Duration(LifeCycle) * time.Second) // never timeout if any message come

			if time.Now().Unix()-clt.lastHandshake < LifeCycle {
				continue
			}

			DelUserClt(clt.User, clt)

			// wait 1s to receive all channel message
			time.Sleep(1 * time.Second)
			close(clt.close)

			return
		}
	}(clt)

	// listen for client message
	go func(clt *Client) {
		for {
			select {
			case msg := <-clt.receiveMessage:
				clt.messages = append(clt.messages, msg)

			case <-clt.fetchMessages:
				clt.msgs <- clt.messages
				clt.messages = make([]*Message, 0, 10)

			case <-clt.close:
				clt.User = nil
				close(clt.receiveMessage)
				close(clt.fetchMessages)
				close(clt.msgs)

				glog.V(2).Infof("clt %s del, break listen loop\n", clt.Id)
				return
			}
		}
	}(clt)

	glog.V(1).Infof("new client %s to user %s\n", clt.Id, clt.User.Id)
	return clt
}

type Client struct {
	Id             string
	User           *User
	receiveMessage chan *Message
	messages       []*Message
	fetchMessages  chan struct{}
	close          chan struct{}
	msgs           chan []*Message
	lastHandshake  int64
	lock           sync.RWMutex
}

func (c *Client) Handshake() {
	glog.V(3).Infof("clt %s handshake\n", c.Id)

	c.lastHandshake = time.Now().Unix()
}

func (c *Client) Msgs() []*Message {
	c.fetchMessages <- struct{}{}
	return <-c.msgs
}
