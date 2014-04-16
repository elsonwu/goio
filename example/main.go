package main

import (
	"encoding/json"
	"fmt"
	"github.com/elsonwu/goreal"
	"github.com/go-martini/martini"
	"log"
	"net/http"
	"runtime"
	"time"
)

func main() {
	host := "127.0.0.1:8888"
	runtime.GOMAXPROCS(runtime.NumCPU())

	clients := goreal.GlobalClients()
	rooms := goreal.GlobalRooms()
	users := goreal.GlobalUsers()

	go func() {
		for {
			time.Sleep(3 * time.Second)
			log.Printf("rooms: %d, users: %d, clients: %d \n\n", rooms.Count(), users.Count(), clients.Count())
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
		// res.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1")
	})

	m.Get("/client/:user_id", func(params martini.Params, req *http.Request) (int, string) {
		userId := params["user_id"]
		user := users.Get(userId)
		if user == nil {
			user = goreal.NewUser(userId)
			users.Add(user)
		}

		clt, done := goreal.NewClient()
		if 0 == user.Clients.Count() {
			user.Emit("broadcast", &goreal.Message{
				EventName: "connect",
				CallerId:  user.Id,
			})
		}

		user.Add(clt)
		done <- true

		clt.User.On("join", func(message *goreal.Message) {
			if roomId, ok := message.Data.(string); ok {
				room := rooms.Get(roomId)
				if !room.Has(clt.User.Id) {
					room.Emit("broadcast", &goreal.Message{
						EventName: "join",
						Data:      clt.User.Id,
					})

					clt.User.On("destory", func(message *goreal.Message) {
						room.Emit("broadcast", &goreal.Message{
							EventName: "leave",
							Data:      clt.User.Id,
						})
					})

					room.Add(clt.User)
				}
			}
		})

		clt.User.On("leave", func(message *goreal.Message) {
			if roomId, ok := message.Data.(string); ok {
				room := rooms.Get(roomId)
				if room.Has(clt.User.Id) {
					room.Emit("broadcast", &goreal.Message{
						EventName: "leave",
						Data:      clt.User.Id,
					})
					room.Delete(clt.User.Id)
				}
			}
		})

		clt.User.On("broadcast", func(message *goreal.Message) {
			if message.RoomId == "" {
				for _, room := range *clt.User.Rooms {
					room.Emit("broadcast", message)
				}
			} else {
				rooms.Get(message.RoomId).Emit("broadcast", message)
			}
		})

		js, _ := json.Marshal(goreal.Message{
			Data:     clt.Id,
			CallerId: userId,
		})
		return 200, string(js)
	})

	m.Get("/message/:id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["id"]
		clt := clients.Get(id)
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

		clt.UpdateActiveTime()
		return 200, "1"
	})

	m.Post("/message/:id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["id"]
		clt := clients.Get(id)
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
			if message.RoomId == "" && (message.EventName == "leave" || message.EventName == "leave") {
				clt.Receive(&goreal.Message{
					EventName: "error",
					Data: map[string]string{
						"error": "room id is missing",
					},
				})
			}

			clt.User.Emit(message.EventName, message)
		}(message)

		clt.UpdateActiveTime()
		return 200, "1"
	})

	log.Println("Serve at " + host)
	if err := http.ListenAndServe(host, m); err != nil {
		log.Println(err)
	}
}
