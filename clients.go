package goio

import (
	"log"
	"sync"
)

func NewClients() *clients {
	clts := new(clients)
	return clts
}

type clients struct {
	m     sync.Map
	count int
}

func (c *clients) Count() int {
	return c.count
}

func (c *clients) addMessage(msg *Message) {
	log.Println("client(s) message " + msg.EventName + " clientId " + msg.ClientId + " callerId " + msg.CallerId)
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
	c.count += 1
	c.m.Store(clt.Id, clt)
}

func (c *clients) DelClt(clt *Client) {
	c.count -= 1
	c.m.Delete(clt.Id)
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
