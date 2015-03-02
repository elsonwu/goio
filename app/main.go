package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/elsonwu/goio"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

var flagAllowOrigin = flag.String("alloworigin", "", "the host allow to cross site ajax")
var flagDebug = flag.Bool("debug", false, "enable debug mode or not")
var flagClientLifeCycle = flag.Int64("lifecycle", 30, "how many seconds of the client life cycle")
var flagClientMessageTimeout = flag.Int64("messagetimeout", 15, "how many seconds of the client keep waiting for new messages")

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

	goji.Abandon(middleware.RequestID)
	goji.Abandon(middleware.Logger)
	goji.Use(func(h http.Handler) http.Handler {
		fn := func(res http.ResponseWriter, r *http.Request) {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			res.Header().Set("Access-Control-Allow-Credentials", "true")
			res.Header().Set("Access-Control-Allow-Methods", "GET,POST")
			if "" != *flagAllowOrigin {
				allowOrigins := strings.Split(*flagAllowOrigin, ",")
				for _, allowOrigin := range allowOrigins {
					res.Header().Add("Access-Control-Allow-Origin", allowOrigin)
				}
			}
			res.Header().Set("Content-Type", "text/plain")
			h.ServeHTTP(res, r)
		}

		return http.HandlerFunc(fn)
	})

	if *flagDebug {
		goji.Get("/test", func(w http.ResponseWriter, req *http.Request) {
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

			io.WriteString(w, strconv.Itoa(int(time.Now().Unix()-st))+" seconds")
		})
	}

	goji.Get("/count", func(w http.ResponseWriter, req *http.Request) {
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

		io.WriteString(w, res)
	})

	goji.Get("/room/users/:room_id", func(ctx web.C, w http.ResponseWriter, req *http.Request) {
		roomId := ctx.URLParams["room_id"]
		if roomId == "" {
			http.Error(w, "room_id is missing", http.StatusForbidden)
			return
		}

		room := rooms.Get(roomId, false)
		if room == nil {
			io.WriteString(w, "room does not exist")
			return
		}

		io.WriteString(w, strings.Join(room.UserIds.Array(), ","))
	})

	goji.Get("/room/users/:room_id", func(ctx web.C, w http.ResponseWriter, req *http.Request) {
		roomId := ctx.URLParams["room_id"]
		if roomId == "" {
			http.Error(w, "room_id is missing", http.StatusForbidden)
			return
		}

		room := rooms.Get(roomId, false)
		if room == nil {
			io.WriteString(w, "")
			return
		}

		io.WriteString(w, strings.Join(room.UserIds.Array(), ","))
	})

	goji.Get("/user/data/:user_id/:key", func(ctx web.C, w http.ResponseWriter, req *http.Request) {
		userId := ctx.URLParams["user_id"]
		if userId == "" {
			http.Error(w, "user_id is missing", http.StatusForbidden)
			return
		}

		key := ctx.URLParams["key"]
		if key == "" {
			http.Error(w, "key is missing", http.StatusForbidden)
			return
		}

		user := users.Get(userId)
		if user == nil {
			io.WriteString(w, "")
			return
		}

		io.WriteString(w, user.Data().Get(key))
	})

	goji.Post("/user/data/:client_id/:key", func(ctx web.C, w http.ResponseWriter, req *http.Request) {
		val, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		clientId := ctx.URLParams["client_id"]
		if clientId == "" {
			http.Error(w, "client_id is missing", http.StatusForbidden)
			return
		}

		key := ctx.URLParams["key"]
		if key == "" {
			http.Error(w, "key is missing", http.StatusForbidden)
			return
		}

		clt := clients.Get(clientId)
		if clt != nil && clt.UserId != "" {
			user := users.Get(clt.UserId)
			if user != nil {
				user.Data().Set(key, string(val))
			}
		}

		io.WriteString(w, "")
	})

	goji.Post("/online_status", func(ctx web.C, w http.ResponseWriter, req *http.Request) {
		val, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
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

		io.WriteString(w, status)
	})

	goji.Post("/client/:user_id", func(ctx web.C, w http.ResponseWriter, req *http.Request) {
		userId := ctx.URLParams["user_id"]
		user := users.Get(userId)
		if user == nil {
			user = goio.NewUser(userId)
		}

		clt, done := goio.NewClient()
		user.Add(clt)
		done <- true

		io.WriteString(w, clt.Id)
	})

	goji.Get("/kill_client/:client_id", func(ctx web.C, w http.ResponseWriter, req *http.Request) {
		id := ctx.URLParams["client_id"]
		clt := clients.Get(id)
		if clt != nil {
			clt.Destroy()
		}

		io.WriteString(w, "")
	})

	goji.Get("/message/:client_id", func(ctx web.C, w http.ResponseWriter, req *http.Request) {
		id := ctx.URLParams["client_id"]
		clt := clients.Get(id)
		if clt == nil {
			http.Error(w, fmt.Sprintf("Client %s does not exist\n", id), http.StatusNotFound)
			return
		}

		clt.Handshake()
		// we handshake again after it finished no matter timeout or ok
		defer clt.Handshake()

		timeNow := time.Now().Unix()
		for {
			if 0 < len(clt.Messages) {
				msgs, err := json.Marshal(clt.Messages)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				clt.CleanMessages()
				io.WriteString(w, string(msgs))
				return
			}

			if time.Now().Unix() > timeNow+*flagClientMessageTimeout {
				break
			}

			time.Sleep(100000 * time.Microsecond)
		}

		io.WriteString(w, "")
	})

	goji.Post("/message/:client_id", func(ctx web.C, w http.ResponseWriter, req *http.Request) {
		id := ctx.URLParams["client_id"]
		clt := clients.Get(id)
		if clt == nil {
			http.Error(w, fmt.Sprintf("Client %s does not exist\n", id), http.StatusForbidden)
			return
		}

		user := users.Get(clt.UserId)
		if user == nil {
			http.Error(w, fmt.Sprintf("Client %s does not connect with any user\n", id), http.StatusForbidden)
			return
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

		io.WriteString(w, "")
	})

	goji.Serve()
}
