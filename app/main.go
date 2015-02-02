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
	goio.Debug = *flagDebug
	goio.LifeCycle = *flagClientLifeCycle

	clients := goio.GlobalClients()
	rooms := goio.GlobalRooms()
	users := goio.GlobalUsers()

	if *flagDebug {
		go func() {
			for {
				time.Sleep(3 * time.Second)
				log.Printf("rooms: %d, users: %d, clients: %d \n", rooms.Count(), users.Count(), clients.Count())
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

			st := time.Now().Unix()
			for i := 0; i < 10000; i++ {
				userId := strconv.Itoa(i)
				user := users.Get(userId)
				if user == nil {
					user = goio.NewUser(userId)
				}

				clt, done := goio.NewClient()
				user.Add(clt)
				done <- true

				room := rooms.Get(strconv.Itoa(i%1000), true)
				room.Add(user)
			}

			return strconv.Itoa(int(time.Now().Unix()-st)) + " seconds"
		})
	}

	m.Get("/count", func(req *http.Request) string {
		res := ""
		res += fmt.Sprintf("rooms: %d, users: %d, clients: %d \n", rooms.Count(), users.Count(), clients.Count())

		if "1" == req.URL.Query().Get("detail") {
			res += fmt.Sprintf("-------------------------------\n")

			rooms.Each(func(room *goio.Room) {
				res += fmt.Sprintf("# room id: %s \n", room.Id)
				for userId, _ := range room.UserIds.Map {
					res += fmt.Sprintf(" - user id: %s \n", userId)
				}
				res += fmt.Sprintf("\n")
			})

			res += fmt.Sprintf("-------------------------------\n")

			users.Each(func(user *goio.User) {
				res += fmt.Sprintf("# user id: %s \n", user.Id)
				for clientId, _ := range user.ClientIds.Map {
					res += fmt.Sprintf(" - client id: %s \n", clientId)
				}
				res += fmt.Sprintf("\n")
			})
		}

		return res
	})

	getRoomUsersFn := func(params martini.Params, req *http.Request) (int, string) {
		roomId := params["room_id"]
		if roomId == "" {
			return 403, "room_id is missing"
		}

		room := rooms.Get(roomId, false)
		if room == nil {
			return 200, ""
		}

		return 200, strings.Join(room.UserIds.Array(), ",")
	}

	// alias of /room/users, will be removed later
	m.Get("/room_users/:room_id", getRoomUsersFn)
	m.Get("/room/users/:room_id", getRoomUsersFn)

	m.Get("/room/data/:room_id/:key", func(params martini.Params, req *http.Request) (int, string) {
		roomId := params["room_id"]
		if roomId == "" {
			return 403, "room_id is missing"
		}

		key := params["key"]
		if key == "" {
			return 403, "key is missing"
		}

		room := rooms.Get(roomId, false)
		if room == nil {
			return 404, "room does not exist"
		}

		return 200, room.Data().Get(key)
	})

	m.Post("/room/data/:room_id/:key", func(params martini.Params, req *http.Request) (int, string) {
		val, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return 500, err.Error()
		}

		roomId := params["room_id"]
		if roomId == "" {
			return 403, "room_id is missing"
		}

		key := params["key"]
		if key == "" {
			return 403, "key is missing"
		}

		room := rooms.Get(roomId, false)
		if room == nil {
			return 404, "room does not exist"
		}

		room.Data().Set(key, string(val))
		return 200, ""
	})

	getUserDataFn := func(params martini.Params, req *http.Request) (int, string) {
		userId := params["user_id"]
		if userId == "" {
			return 403, "user_id is missing"
		}

		key := params["key"]
		if key == "" {
			return 403, "key is missing"
		}

		user := users.Get(userId)
		if user == nil {
			return 200, ""
		}

		return 200, user.Data().Get(key)
	}

	m.Get("/user/data/:user_id/:key", getUserDataFn)
	// alias of /user/data, will be removed later
	m.Get("/get_data/:user_id/:key", getUserDataFn)

	setUserDataFn := func(params martini.Params, req *http.Request) (int, string) {
		val, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return 500, err.Error()
		}

		clientId := params["client_id"]
		if clientId == "" {
			return 403, "client_id is missing"
		}

		key := params["key"]
		if key == "" {
			return 403, "key is missing"
		}

		clt := clients.Get(clientId)
		if clt != nil && clt.UserId != "" {
			user := users.Get(clt.UserId)
			if user != nil {
				user.Data().Set(key, string(val))
			}
		}

		return 200, ""
	}

	m.Post("/user/data/:client_id/:key", setUserDataFn)
	// alias of /user/data, will be removed later
	m.Post("/set_data/:client_id/:key", setUserDataFn)

	m.Post("/online_status", func(params martini.Params, req *http.Request) (int, string) {
		val, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return 500, err.Error()
		}

		userIds := strings.Split(string(val), ",")
		status := ""
		for _, userId := range userIds {
			if nil == users.Get(userId) {
				status += "0,"
			} else {
				status += "1,"
			}
		}

		return 200, status
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

	m.Get("/kill_client/:client_id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["client_id"]
		clt := clients.Get(id)
		if clt != nil {
			clt.Destroy()
		}

		return 204, ""
	})

	m.Get("/message/:client_id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["client_id"]
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
				msgs, err := json.Marshal(clt.Messages)
				if err != nil {
					return 500, err.Error()
				}

				clt.CleanMessages()
				return 200, string(msgs)
			}

			if time.Now().Unix() > timeNow+*flagClientMessageTimeout {
				break
			}

			time.Sleep(100000 * time.Microsecond)
		}

		return 204, ""
	})

	m.Post("/message/:client_id", func(params martini.Params, req *http.Request) (int, string) {
		id := params["client_id"]
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
		defer req.Body.Close()

		message := &goio.Message{}
		json.NewDecoder(req.Body).Decode(message)
		message.CallerId = clt.UserId
		go func(message *goio.Message) {
			if *flagDebug {
				log.Printf("post message: %#v", *message)
			}

			if message.RoomId == "" && (message.EventName == "join" || message.EventName == "leave") {
				clt.Receive(&goio.Message{
					EventName: "error",
					Data:      "room id is missing",
				})

				return
			}

			// We change CallerId as current user
			message.CallerId = user.Id
			message.ClientId = clt.Id
			user.Emit(message.EventName, message)
		}(message)

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
