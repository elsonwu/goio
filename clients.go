package goio

import (
	"sync"

	"github.com/golang/glog"
)

func NewClients() *clients {
	clts := new(clients)
	return clts
}

type clients struct {
	m sync.Map
}

func (c *clients) Count() int {
	n := 0
	c.Range(func(c *Client) {
		n = n + 1
	})

	return n
}

func (c *clients) addMessage(msg *Message) {
	glog.V(1).Infoln("client(s) message " + msg.EventName + " clientId " + msg.ClientId + " callerId " + msg.CallerId)
	c.m.Range(func(k interface{}, v interface{}) bool {
		clt := v.(*Client)
		if clt == nil || clt.IsDead() || clt.Id == msg.ClientId {
			return true
		}

		clt.addMessage(msg)
		return true
	})
}

func (c *clients) AddClt(clt *Client) {
	c.m.Store(clt.Id, clt)
}

func (c *clients) DelClt(clientId string) {
	c.m.Delete(clientId)
}

func (c *clients) Get(clientId string) *Client {
	v, ok := c.m.Load(clientId)
	if !ok {
		return nil
	}

	return v.(*Client)
}

func (c *clients) Range(f func(r *Client)) {
	c.m.Range(func(k interface{}, v interface{}) bool {
		client, ok := v.(*Client)
		if !ok {
			return true
		}

		f(client)
		return true
	})
}
