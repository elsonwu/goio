package goio

import "sync"

func NewClients() *clients {
	clts := new(clients)
	clts.Clients = make(map[string]*Client)
	clts.Message = make(chan *Message)
	clts.AddClt = make(chan *Client)
	clts.DelClt = make(chan *Client)

	clts.clt = make(chan *Client)
	clts.getClt = make(chan string)

	clts.getCount = make(chan struct{})
	clts.count = make(chan int)

	go func(clts *clients) {
		for {
			select {
			case c := <-clts.AddClt:
				clts.Clients[c.Id] = c

			case c := <-clts.DelClt:
				delete(clts.Clients, c.Id)

			case clientId := <-clts.getClt:
				client, _ := clts.Clients[clientId]
				clts.clt <- client

			case <-clts.getCount:
				clts.count <- len(clts.Clients)
			}
		}

	}(clts)

	return clts
}

type clients struct {
	Clients map[string]*Client
	Message chan *Message
	AddClt  chan *Client
	DelClt  chan *Client

	clt    chan *Client
	getClt chan string

	count    chan int
	getCount chan struct{}
	lock     sync.RWMutex
}

func (c *clients) Count() int {
	c.getCount <- struct{}{}
	return <-c.count
}

func (c *clients) Get(clientId string) *Client {
	c.getClt <- clientId
	return <-c.clt
}
