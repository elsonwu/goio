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
			log.Printf("# count room: %d, clients: %d # \n", len(crh), len(ch))
		}
	}()

	martini.Env = martini.Dev
	router := martini.NewRouter()
	mart := martini.New()
	mart.Action(router.Handle)
	m := &martini.ClassicMartini{mart, router}
	m.Use(martini.Recovery())
	m.Use(func(res http.ResponseWriter) {
		res.Header().Set("Content-Type", "application/json")
		res.Header().Set("Access-Control-Allow-Methods", "GET, POST")
		res.Header().Set("Access-Control-Allow-Credentials", "true")
		res.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1")
	})

	m.Get("/client", func(req *http.Request) (int, string) {
		id := uuid()
		clt := ch.Add(id)
		clt.Emit("broadcast", &goreal.Message{
			EventName: "connect",
			Data:      id,
		})

		clt.On("destory", func(message *goreal.Message) {
			clt.Emit("broadcast", &goreal.Message{
				EventName: "disconnect",
				Data:      message.Data,
			})
		})

		clt.On("join", func(message *goreal.Message) {
			if roomId, ok := message.Data.(string); ok {
				room := crh.Room(roomId)
				if !room.Has(clt.Id) {
					room.Emit("broadcast", &goreal.Message{
						EventName: "joined",
						Data:      clt.Id,
					})
					room.Add(clt)
				}
			}
		})

		clt.On("leave", func(message *goreal.Message) {
			if roomId, ok := message.Data.(string); ok {
				room := crh.Room(roomId)
				if room.Has(clt.Id) {
					room.Emit("broadcast", &goreal.Message{
						EventName: "left",
						Data:      clt.Id,
					})
					room.Delete(clt.Id)
				}
			}
		})

		return 200, id
	})

	m.Get("/message/:id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["id"]
		clt := ch.Client(id)
		if clt == nil {
			return 404, fmt.Sprintf("Client %s does not exist\n", id)
		}

		clt.UpdateActiveTime()
		select {
		case msg := <-clt.Msg:
			js, _ := json.Marshal(msg)
			return 200, string(js)
		case <-time.After(10 * time.Second):
			// do nothing
		}

		return 200, "1"
	})

	m.Post("/message/:id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["id"]
		clt := ch.Client(id)
		if clt == nil {
			return 404, fmt.Sprintf("Client %s does not exist\n", id)
		}

		clt.UpdateActiveTime()
		message := &goreal.Message{}
		defer req.Body.Close()
		err := json.NewDecoder(req.Body).Decode(message)
		if err != nil {
			return 200, "message format is invalid"
		}

		go func(message *goreal.Message) {
			log.Println("post message: ", *message)
			clt.Emit(message.EventName, message)
		}(message)

		return 200, "1"
	})

	log.Println("Serve at " + host)
	if err := http.ListenAndServe(host, m); err != nil {
		log.Println(err)
	}
}
