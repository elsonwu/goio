package goreal

import (
	"fmt"
	"time"
)

type ClientHandler map[string]*Client

func (self *ClientHandler) Client(id string) *Client {
	if clt, ok := (*self)[id]; ok {
		return clt
	}

	return nil
}

func (self *ClientHandler) Delete(id string) {
	delete(*self, id)
}

func (self *ClientHandler) Add(id string) *Client {
	clt := &Client{
		Id:            id,
		Msg:           make(chan *Message),
		LastHandshake: time.Now().Unix(),
	}

	(*self)[id] = clt

	clt.On("destory", func(message *Message) {
		self.Delete(id)
	})

	go func(id string) {
		for {
			clt := self.Client(id)
			if clt == nil {
				break
			}

			time.Sleep(3 * time.Second)
			fmt.Println("id:", clt.Id, " last:", clt.LastHandshake, " t:", time.Now().Unix()-clt.LastHandshake)
			if 30 < time.Now().Unix()-clt.LastHandshake {
				fmt.Println("client id:", clt.Id, " destory")
				clt.Destory()
			}
		}
	}(id)

	return clt
}
