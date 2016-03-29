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
var flagAllowOrigin = flag.String("alloworigin", "*", "the host allow to cross site ajax")
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

	goio.LifeCycle = *flagClientLifeCycle

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
		res.Header().Set("Powered-By", "Goio")
		if "" != *flagAllowOrigin {
			allowOrigins := strings.Split(*flagAllowOrigin, ",")
			for _, allowOrigin := range allowOrigins {
				res.Header().Add("Access-Control-Allow-Origin", allowOrigin)
			}
		}
	})

	if *flagDebug {

		m.Get("/test/message", func() string {
			userId1 := "u1"
			user1 := goio.Users().MustGet(userId1)
			clt1 := goio.NewClient(user1)

			roomId := "r1"
			room := goio.Rooms().Get(roomId)
			if room == nil {
				room = goio.NewRoom(roomId)
			}

			room.AddUser <- user1

			go func(clt1 *goio.Client) {
				for {
					msgs := clt1.Msgs()
					if len(msgs) == 0 {
						time.Sleep(3 * time.Second)
						continue
					}

					fmt.Printf("user %s clt %s received message %#v \n", clt1.User.Id, clt1.Id, msgs)
				}
			}(clt1)

			userId2 := "u2"
			user2 := goio.Users().MustGet(userId2)
			clt2 := goio.NewClient(user2)

			msg := &goio.Message{}
			msg.CallerId = userId2
			msg.ClientId = clt2.Id
			msg.EventName = "join"
			msg.RoomId = roomId
			msg.Data = `{"val":"this is a test"}`

			room.Message <- msg

			return "completed"
		})

		m.Get("/test/client", func() string {

			count := 10000
			all := make(chan struct{}, count)

			st := time.Now().Second()
			for i := 1; i <= count; i++ {
				go func(i int) {
					userId := strconv.Itoa(i)
					user := goio.Users().MustGet(userId)

					goio.NewClient(user)

					roomId := strconv.Itoa(i % 1000)
					room := goio.Rooms().Get(roomId)
					if room == nil {
						room = goio.NewRoom(roomId)
					}

					room.AddUser <- user

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

	// get user ids array from the given room
	m.Get("/room/users/:room_id", func(params martini.Params, req *http.Request) (int, string) {
		roomId := params["room_id"]

		room := goio.Rooms().Get(roomId)
		if room == nil {
			return 200, ""
		}

		return 200, strings.Join(room.UserIds(), ",")
	})

	m.Get("/user/data/:user_id/:key", func(params martini.Params, req *http.Request) (int, string) {
		userId := params["user_id"]
		key := params["key"]

		user := goio.Users().Get(userId)
		if user == nil {
			return 200, ""
		}

		return 200, user.GetData(key)
	})

	m.Post("/user/data/:client_id/:key", func(params martini.Params, req *http.Request) (int, string) {
		val, err := ioutil.ReadAll(req.Body)
		req.Body.Close()

		if err != nil {
			return 500, err.Error()
		}

		clientId := params["client_id"]
		key := params["key"]

		clt := goio.Clients().Get(clientId)
		if clt != nil && clt.User != nil {
			clt.User.AddData <- goio.UserData{Key: key, Val: string(val)}
		}

		return 200, ""
	})

	m.Post("/online_status", func(params martini.Params, req *http.Request) (int, string) {
		val, err := ioutil.ReadAll(req.Body)
		req.Body.Close()

		if err != nil {
			fmt.Printf("online_status error %s\n", err.Error())
			return 500, err.Error()
		}

		userIds := strings.Split(string(val), ",")
		fmt.Printf("checking userIds:%s\n", string(val))

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
		user := goio.Users().MustGet(userId)
		clt := goio.NewClient(user)

		return 200, clt.Id
	})

	m.Get("/kill_client/:client_id", func(params martini.Params, req *http.Request) (int, string) {

		clt := goio.Clients().Get(params["client_id"])
		if clt != nil {
			clt.User.DelClt <- clt
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

		for {
			msgs := clt.Msgs()
			if len(msgs) == 0 {
				time.Sleep(3 * time.Second)
				continue
			}

			m, err := json.Marshal(msgs)
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

		msg := &goio.Message{}
		json.NewDecoder(req.Body).Decode(msg)
		req.Body.Close()

		msg.CallerId = clt.User.Id
		msg.ClientId = clt.Id

		go func(msg *goio.Message, clt *goio.Client) {
			if *flagDebug {
				log.Printf("post msg: %#v\n", *msg)
			}

			fmt.Println("user send msg - begin")

			switch msg.EventName {
			case "join":
				fmt.Printf("msg event - join\n")
				if msg.RoomId != "" {
					r := goio.Rooms().Get(msg.RoomId)
					if r == nil {
						r = goio.NewRoom(msg.RoomId)
					}

					r.AddUser <- clt.User
					r.Message <- msg
				}

			case "leave":
				fmt.Printf("msg event - leave\n")
				if msg.RoomId != "" {
					r := goio.Rooms().Get(msg.RoomId)
					if r != nil {
						r.DelUser <- clt.User
						r.Message <- msg
					}
				}
			case "broadcast":
				fmt.Printf("msg event - broadcast\n")
				if msg.RoomId != "" {
					r := goio.Rooms().Get(msg.RoomId)
					if r != nil {
						r.Message <- msg
					}
				} else {
					for _, r := range clt.User.Rooms() {
						r.Message <- msg
					}
				}

			default:
				fmt.Println("unknown user msg")
			}

			fmt.Println("user send msg - done")

		}(msg, clt)

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
