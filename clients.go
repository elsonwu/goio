package goio

import "sync"

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

func (c *clients) AddMessage(msg *Message) {
	c.m.Range(func(k interface{}, v interface{}) bool {
		clt := v.(*Client)
		if clt.died {
			return true
		}

		go clt.AddMessage(msg)
		return true
	})
}

func (c *clients) AddClt(clt *Client) {
	if clt.died {
		return
	}

	c.count = c.count + 1
	c.m.Store(clt.Id, clt)
}

func (c *clients) DelClt(clt *Client) {
	c.count = c.count - 1
	c.m.Delete(clt.Id)
}

func (c *clients) Get(clientId string) *Client {
	v, ok := c.m.Load(clientId)
	if !ok {
		return nil
	}

	return v.(*Client)
}
