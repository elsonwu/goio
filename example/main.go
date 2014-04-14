package main

import (
	"encoding/json"
	"fmt"
	"github.com/elsonwu/goreal"
	"github.com/go-martini/martini"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

func uuid() string {
	return strconv.Itoa(int(time.Now().Nanosecond()))
}

func main() {
	host := "127.0.0.1:8888"
	runtime.GOMAXPROCS(runtime.NumCPU())

	ch := goreal.NewClientHandler()
	crh := goreal.NewClientRoomHandler()

	go func() {
		for {
			time.Sleep(10 * time.Second)
			fmt.Printf("clients: %#v\n", ch)
			for _, r := range crh {
				fmt.Printf("id: %s, clients: %#v\n", r.Id, r.Clients)
			}
		}
	}()

	martini.Env = martini.Dev
	router := martini.NewRouter()
	mart := martini.New()
	mart.Action(router.Handle)
	m := &martini.ClassicMartini{mart, router}
	m.Use(martini.Recovery())
	m.Use(func(res http.ResponseWriter) {
		res.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		res.Header().Set("Content-Type", "application/json")
	})

	m.Get("/new_client", func(req *http.Request) (int, string) {
		id := uuid()
		client := ch.Add(id)
		client.Emit("connect", &goreal.Message{
			EventName: "connect",
			Data:      id,
		})

		client.On("broadcast", func(message *goreal.Message) {
			fmt.Println("id:"+id+" received message ", *message)
		})

		return 200, id
	})

	m.Get("/join/:id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["id"]
		clt := ch.Client(id)
		if clt == nil {
			return 403, fmt.Sprintf("client %s does not exist\n", id)
		}

		clt.UpdateActiveTime()
		roomId := req.URL.Query().Get("room_id")
		if roomId == "" {
			return 403, fmt.Sprintf("room_id is missing\n")
		}

		go func() {
			roomId := req.URL.Query().Get("room_id")
			room := crh.Room(roomId)
			if !room.Has(clt.Id) {
				room.Emit("broadcast", &goreal.Message{
					EventName: "joined",
					Data:      clt.Id,
				})
				room.Add(clt)
			}
		}()

		return 200, "ok"
	})

	m.Get("/message/:id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["id"]
		clt := ch.Client(id)
		if clt == nil {
			return 401, fmt.Sprintf("To client %s does not exist\n", id)
		}

		clt.UpdateActiveTime()
		select {
		case msg := <-clt.Msg:
			fmt.Println("msg:", msg)
			js, _ := json.Marshal(msg)
			return 200, string(js)
		case <-time.After(10 * time.Second):
			fmt.Println("no msg...")
			return 200, "1"
		}

		return 200, "1"
	})

	m.Post("/message/:id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["id"]
		clt := ch.Client(id)
		if clt == nil {
			return 401, fmt.Sprintf("To client %s does not exist\n", id)
		}

		clt.UpdateActiveTime()
		go func() {
			message := &goreal.Message{}
			err := json.NewDecoder(req.Body).Decode(message)
			if err != nil {
				return
			}

			clt.Emit("broadcast", message)
		}()

		return 200, "1"
	})

	log.Println("Serve at " + host)
	if err := http.ListenAndServe(host, m); err != nil {
		log.Println(err)
	}
}
