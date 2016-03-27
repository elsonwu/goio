package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/elsonwu/goio"
	"github.com/go-martini/martini"
)

var flagHost = flag.String("host", "127.0.0.1:9999", "the default server host")
var flagSSLHost = flag.String("sslhost", "", "the server host for https, it will override the host setting")
var flagAllowOrigin = flag.String("alloworigin", "", "the host allow to cross site ajax")
var flagDebug = flag.Bool("debug", false, "enable debug mode or not")
var flagClientLifeCycle = flag.Int64("lifecycle", 30, "how many seconds of the client life cycle")
var flagClientMessageTimeout = flag.Int64("messagetimeout", 15, "how many seconds of the client keep waiting for new messages")
var flagEnableHttps = flag.Bool("enablehttps", false, "enable https or not")
var flagDisableHttp = flag.Bool("disablehttp", false, "disable http and use https only")
var flagCertFile = flag.String("certfile", "", "certificate file path")
var flagKeyFile = flag.String("keyfile", "", "private file path")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	if *flagDebug {
		go func() {
			for {
				time.Sleep(1 * time.Second)
				log.Printf("rooms: %d, users: %d, clients: %d \n", goio.Rooms().Count(), goio.Users().Count(), goio.Clients().Count())
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
		res.Header().Set("Access-Control-Allow-Credentials", "true")
		res.Header().Set("Access-Control-Allow-Methods", "GET,POST")
		if "" != *flagAllowOrigin {
			allowOrigins := strings.Split(*flagAllowOrigin, ",")
			for _, allowOrigin := range allowOrigins {
				res.Header().Add("Access-Control-Allow-Origin", allowOrigin)
			}
		}
	})

	if *flagDebug {
		m.Get("/test", func() string {

			count := 10000
			all := make(chan struct{}, count)

			st := time.Now().Second()
			for i := 1; i <= count; i++ {
				go func(i int) {
					userId := strconv.Itoa(i)
					user := goio.Users().Get(userId)
					if user == nil {
						user = goio.NewUser(userId)
					}
					goio.Users().AddUser <- user

					clt := goio.NewClient(user)
					user.AddClt <- clt
					goio.Clients().AddClt <- clt

					roomId := strconv.Itoa(i % 1000)
					room := goio.Rooms().Get(roomId)
					if room == nil {
						room = goio.NewRoom(roomId)
					}

					room.AddUser <- user
					goio.Rooms().AddRoom <- room

					all <- struct{}{}
				}(i)
			}

			for i := 0; i < count; i++ {
				<-all
			}

			return fmt.Sprintf("%d s", time.Now().Second()-st)
		})
	}

	m.Get("/count", func(req *http.Request) string {
		return fmt.Sprintf("rooms: %d, users: %d, clients: %d \n", goio.Rooms().Count(), goio.Users().Count(), goio.Clients().Count())
	})

	m.Get("/room/users/:room_id", func(params martini.Params, req *http.Request) (int, string) {
		roomId := params["room_id"]

		room := goio.Rooms().Get(roomId)
		if room == nil {
			return 200, ""
		}

		return 200, strings.Join(room.UserIds(), ",")
	})

	// m.Get("/user/data/:user_id/:key", func(params martini.Params, req *http.Request) (int, string) {
	// userId := params["user_id"]
	// key := params["key"]

	// user := goio.Users().Get(userId)
	// if user == nil {
	// return 200, ""
	// }

	// return 200, user.Data().Get(key)
	// })

	// m.Post("/user/data/:client_id/:key", func(params martini.Params, req *http.Request) (int, string) {
	// val, err := ioutil.ReadAll(req.Body)
	// if err != nil {
	// return 500, err.Error()
	// }

	// clientId := params["client_id"]
	// key := params["key"]

	// clt := clients.Get(clientId)
	// if clt != nil && clt.User != nil {
	// clt.User.Data().Set(key, string(val))
	// }

	// return 200, ""
	// })

	m.Post("/online_status", func(params martini.Params, req *http.Request) (int, string) {
		val, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return 500, err.Error()
		}

		userIds := strings.Split(string(val), ",")
		status := ""
		for _, userId := range userIds {
			if nil == goio.Users().Get(userId) {
				status += "0,"
			} else {
				status += "1,"
			}
		}

		return 200, status
	})

	m.Post("/client/:user_id", func(params martini.Params, req *http.Request) (int, string) {
		userId := params["user_id"]
		user := goio.Users().Get(userId)
		if user == nil {
			user = goio.NewUser(userId)
		}

		clt := goio.NewClient(user)
		return 200, clt.Id
	})

	m.Get("/kill_client/:client_id", func(params martini.Params, req *http.Request) (int, string) {

		clt := goio.Clients().Get(params["client_id"])
		if clt != nil {
			goio.Clients().DelClt <- clt
		}

		return 204, ""
	})

	m.Get("/message/:client_id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["client_id"]
		clt := goio.Clients().Get(id)

		if clt == nil {
			return 404, fmt.Sprintf("Client %s does not exist\n", id)
		}

		clt.Handshake()
		// we handshake again after it finished no matter timeout or ok
		defer clt.Handshake()

		select {
		case <-time.After(time.Duration(*flagClientMessageTimeout) * time.Second):
			return 204, ""
		case msg := <-clt.Message:
			m, err := json.Marshal(msg)
			if err != nil {
				return 500, err.Error()
			}

			return 200, string(m)
		}
	})

	m.Post("/message/:client_id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["client_id"]
		clt := goio.Clients().Get(id)
		if clt == nil {
			return 403, fmt.Sprintf("Client %s does not exist\n", id)
		}

		if clt.User == nil {
			return 403, fmt.Sprintf("Client %s does not connect with any user\n", id)
		}

		defer req.Body.Close()

		message := &goio.Message{}
		json.NewDecoder(req.Body).Decode(message)
		message.CallerId = clt.User.Id

		go func(message *goio.Message, clt *goio.Client) {
			if *flagDebug {
				log.Printf("post message: %#v", *message)
			}

			// We change CallerId as current user
			message.CallerId = clt.User.Id
			message.ClientId = clt.Id
			switch message.EventName {
			case "broadcast":
				if message.RoomId == "" {
					room := goio.Rooms().Get(message.RoomId)
					if room == nil {
						clt.Message <- &goio.Message{
							EventName: "error",
							Data:      "room id is invalid",
						}

						return
					}

					room.Message <- message
				} else {
					goio.Clients().Message <- message
				}

			case "join":
				if message.RoomId == "" {
					clt.Message <- &goio.Message{
						EventName: "error",
						Data:      "room id is missing",
					}

					return
				}

				room := goio.Rooms().Get(message.RoomId)
				if room == nil {
					clt.Message <- &goio.Message{
						EventName: "error",
						Data:      "room id is invalid",
					}

					return
				}

				room.AddUser <- clt.User

			case "leave":
				if message.RoomId == "" {
					clt.Message <- &goio.Message{
						EventName: "error",
						Data:      "room id is missing",
					}

					return
				}

				room := goio.Rooms().Get(message.RoomId)
				if room == nil {
					clt.Message <- &goio.Message{
						EventName: "error",
						Data:      "room id is invalid",
					}

					return
				}

				room.DelUser <- clt.User
			}
		}(message, clt)

		return 204, ""
	})

	m.Options("/.*", func(req *http.Request) {})

	host := *flagHost
	if !*flagEnableHttps && *flagDisableHttp {
		log.Fatalln("You cannot disable http but not enable https in the same time")
	}

	//Prevent exiting
	ch := make(chan bool)

	if !*flagDisableHttp {
		go func() {
			log.Println("Serve at " + host + " - http")
			if err := http.ListenAndServe(host, m); err != nil {
				log.Println(err)
			}
		}()
	}

	if *flagEnableHttps {
		go func() {
			if *flagSSLHost != "" {
				host = *flagSSLHost
			}

			log.Println("Serve at " + host + " - https")
			if err := http.ListenAndServeTLS(host, *flagCertFile, *flagKeyFile, m); err != nil {
				log.Println(err)
			}
		}()
	}

	<-ch
}
