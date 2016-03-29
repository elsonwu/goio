package goio

import "sync"

func NewClients() *clients {
	clts := new(clients)
	clts.Clients = make(map[string]*Client)
	clts.Message = make(chan *Message)
	clts.addClt = make(chan *Client)
	clts.delClt = make(chan *Client)

	clts.clt = make(chan *Client)
	clts.getClt = make(chan string)

	clts.getCount = make(chan struct{})
	clts.count = make(chan int)

	go func(clts *clients) {
		for {
			select {
			case c := <-clts.addClt:
				clts.Clients[c.Id] = c

			case c := <-clts.delClt:
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
	addClt  chan *Client
	delClt  chan *Client

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
