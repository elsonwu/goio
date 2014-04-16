package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/elsonwu/goreal"
	"github.com/go-martini/martini"
	"log"
	"net/http"
	"runtime"
	"time"
)

var flagHost = flag.String("host", "127.0.0.1", "the server host")
var flagPort = flag.String("port", "9999", "the server port")
var flagAllowOrigin = flag.String("alloworigin", "", "the host allow to cross site ajax")
var flagDebug = flag.Bool("debug", false, "enable debug mode or not")
var flagClientLifeCycle = flag.Int64("lifecycle", 60, "how many seconds of the client life cycle")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	goreal.Debug = *flagDebug
	goreal.LifeCycle = *flagClientLifeCycle

	clients := goreal.GlobalClients()
	rooms := goreal.GlobalRooms()
	users := goreal.GlobalUsers()

	if *flagDebug {
		go func() {
			for {
				time.Sleep(3 * time.Second)
				log.Printf("rooms: %d, users: %d, clients: %d \n\n", rooms.Count(), users.Count(), clients.Count())
			}
		}()
	}

	martini.Env = martini.Dev
	router := martini.NewRouter()
	mart := martini.New()
	mart.Action(router.Handle)
	m := &martini.ClassicMartini{mart, router}
	m.Use(martini.Recovery())
	m.Use(func(res http.ResponseWriter) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.Header().Set("Access-Control-Allow-Methods", "GET, POST")
		if "" != *flagAllowOrigin {
			res.Header().Set("Access-Control-Allow-Origin", *flagAllowOrigin)
		}
	})

	m.Post("/client/:user_id", func(params martini.Params, req *http.Request) (int, string) {
		userId := params["user_id"]
		user := users.Get(userId)
		if user == nil {
			user = goreal.NewUser(userId)
			users.Add(user)
		}

		clt, done := goreal.NewClient()
		user.Add(clt)
		done <- true

		return 200, clt.Id
	})

	m.Get("/message/:id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["id"]
		clt := clients.Get(id)
		if clt == nil {
			return 404, fmt.Sprintf("Client %s does not exist\n", id)
		}

		clt.Handshake()
		select {
		case msg := <-clt.Msg:
			js, _ := json.Marshal(msg)
			return 200, string(js)
		case <-time.After(30 * time.Second):
			// do nothing
		}

		clt.Handshake()
		return 200, ""
	})

	m.Post("/message/:id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["id"]
		clt := clients.Get(id)
		if clt == nil {
			return 403, fmt.Sprintf("Client %s does not exist\n", id)
		}

		clt.Handshake()
		message := &goreal.Message{}
		defer req.Body.Close()
		err := json.NewDecoder(req.Body).Decode(message)
		if err != nil {
			return 400, "message format is invalid"
		}

		go func(message *goreal.Message) {
			if *flagDebug {
				log.Println("post message: ", *message)
			}

			if message.RoomId == "" && (message.EventName == "join" || message.EventName == "leave") {
				clt.Receive(&goreal.Message{
					EventName: "error",
					Data: map[string]string{
						"error": "room id is missing",
					},
				})
			}

			clt.User.Emit(message.EventName, message)
		}(message)

		clt.Handshake()
		return 200, ""
	})

	host := *flagHost + ":" + *flagPort
	log.Println("Serve at " + host)
	if err := http.ListenAndServe(host, m); err != nil {
		log.Println(err)
	}
}
