package goio

import (
	"log"
	"time"

	"github.com/globalsign/mgo/bson"
)

func newClientId() string {
	return bson.NewObjectId().Hex()
}

func NewClient(user *User) *Client {
	clt := &Client{
		Id:   newClientId(),
		User: user,
		died: false,
	}

	clt.Ping()

	Clients().AddClt(clt)
	user.AddClt(clt)

	return clt
}

type Client struct {
	Id            string
	User          *User
	messages      []*Message
	lastHandshake int64
	died          bool
}

func (c *Client) Ping() {
	c.died = false
	c.lastHandshake = time.Now().Unix()
}

func (c *Client) SetIsDead() {
	c.died = true
}

func (c *Client) IsDead() bool {
	return c.died || c.lastHandshake+LifeCycle < time.Now().Unix()
}

func (c *Client) addMessage(msg *Message) {
	if c.IsDead() {
		return
	}

	log.Println("client.addMessage " + c.Id + " message " + msg.EventName)
	c.messages = append(c.messages, msg)
}

func (c *Client) ReadMessages() []*Message {
	if c.IsDead() {
		return nil
	}

	c.Ping()

	if len(c.messages) == 0 {
		return nil
	}

	msgs := c.messages
	c.messages = make([]*Message, 0, 10)
	return msgs
}
