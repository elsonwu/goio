package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/elsonwu/goio"
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
	goio.Debug = *flagDebug
	goio.LifeCycle = *flagClientLifeCycle

	clients := goio.GlobalClients()
	rooms := goio.GlobalRooms()
	users := goio.GlobalUsers()

	go func() {
		for {
			time.Sleep(3 * time.Second)
			log.Printf("rooms: %d, users: %d, clients: %d \n", rooms.Count(), users.Count(), clients.Count())
		}
	}()

	martini.Env = martini.Dev
	router := martini.NewRouter()
	mart := martini.New()
	mart.Action(router.Handle)
	m := &martini.ClassicMartini{mart, router}
	m.Use(martini.Recovery())
	m.Use(func(res http.ResponseWriter) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.Header().Set("Access-Control-Allow-Credentials", "true")
		res.Header().Set("Access-Control-Allow-Methods", "GET,POST")
		if "" != *flagAllowOrigin {
			res.Header().Set("Access-Control-Allow-Origin", *flagAllowOrigin)
		}
	})

	m.Post("/client/:user_id", func(params martini.Params, req *http.Request) (int, string) {
		userId := params["user_id"]
		user := users.Get(userId)
		if user == nil {
			user = goio.NewUser(userId)
		}

		clt, done := goio.NewClient()
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
		// we handshake again after it finished no matter timeout or ok
		defer clt.Handshake()

		timeNow := time.Now().Unix()
		for {
			if 0 < len(clt.Messages) {
				msg := clt.Messages[0:1]
				clt.Messages = clt.Messages[1:]
				js, err := json.Marshal(msg)
				if err != nil {
					return 500, err.Error()
				}

				return 200, string(js)
			}

			if time.Now().Unix() > timeNow+30 {
				break
			}

			time.Sleep(3 * time.Second)
		}

		return 200, ""
	})

	m.Post("/message/:id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["id"]
		clt := clients.Get(id)
		if clt == nil {
			return 403, fmt.Sprintf("Client %s does not exist\n", id)
		}

		user := users.Get(clt.UserId)
		if user == nil {
			return 403, fmt.Sprintf("Client %s does not connect with any user\n", id)
		}

		clt.Handshake()
		// we handshake again after it finished no matter timeout or ok
		defer clt.Handshake()

		message := &goio.Message{}
		defer req.Body.Close()
		err := json.NewDecoder(req.Body).Decode(message)
		if err != nil {
			return 400, "message format is invalid"
		}

		go func(message *goio.Message) {
			if *flagDebug {
				log.Printf("post message: %#v", *message)
			}

			if message.RoomId == "" && (message.EventName == "join" || message.EventName == "leave") {
				clt.Receive(&goio.Message{
					EventName: "error",
					Data: map[string]string{
						"error": "room id is missing",
					},
				})
			}

			// We change CallerId as current user
			message.CallerId = user.Id
			user.Emit(message.EventName, message)
			log.Printf("post message %#v", *message)
		}(message)

		return 200, ""
	})

	m.Options("/.*", func(req *http.Request) {})

	host := *flagHost + ":" + *flagPort
	log.Println("Serve at " + host)
	if err := http.ListenAndServe(host, m); err != nil {
		log.Println(err)
	}
}
