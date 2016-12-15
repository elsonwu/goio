package goio

func NewClients() *clients {
	clts := new(clients)
	clts.Clients = make(map[string]*Client)
	clts.Message = make(chan *Message)

	clts.addClt = make(chan *Client)
	clts.delClt = make(chan *Client)
	clts.getClt = make(chan cltGetter)
	clts.getCount = make(chan chan int)

	go func(clts *clients) {
		for {
			select {
			case msg := <-clts.Message:
				for _, c := range clts.Clients {
					c.AddMessage(msg)
				}

			case c := <-clts.addClt:
				clts.Clients[c.Id] = c

			case c := <-clts.delClt:
				delete(clts.Clients, c.Id)

			case getter := <-clts.getClt:
				clt, _ := clts.Clients[getter.clientId]
				getter.client <- clt
			case counter := <-clts.getCount:
				counter <- len(clts.Clients)
			}
		}
	}(clts)

	return clts
}

type cltGetter struct {
	clientId string
	client   chan *Client
}

type clients struct {
	Clients map[string]*Client

	Message chan *Message

	addClt   chan *Client
	delClt   chan *Client
	getClt   chan cltGetter
	getCount chan chan int
}

func (c *clients) Count() int {
	counter := make(chan int)
	c.getCount <- counter

	defer close(counter)
	return <-counter
}

func (c *clients) Get(clientId string) *Client {
	clt := make(chan *Client, 1)
	c.getClt <- cltGetter{
		clientId: clientId,
		client:   clt,
	}

	defer close(clt)
	return <-clt
}
