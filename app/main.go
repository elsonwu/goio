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
	"github.com/gin-gonic/gin"
)

var flagHost = flag.String("host", "127.0.0.1:9999", "the default server host")
var flagSSLHost = flag.String("sslhost", "", "the server host for https, it will override the host setting")
var flagAllowOrigin = flag.String("alloworigin", "*", "the host allow to cross site ajax")
var flagDebug = flag.Bool("debug", false, "enable debug mode or not")
var flagClientLifeCycle = flag.Int64("lifecycle", 30, "how many seconds of the client life cycle")
var flagClientMessageTimeout = flag.Int("messagetimeout", 15, "how many seconds of the client keep waiting for new messages")
var flagEnableHttps = flag.Bool("enablehttps", false, "enable https or not")
var flagDisableHttp = flag.Bool("disablehttp", false, "disable http and use https only")
var flagCertFile = flag.String("certfile", "", "certificate file path")
var flagKeyFile = flag.String("keyfile", "", "private file path")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	goio.Rooms()
	goio.Users()
	goio.Clients()

	if *flagDebug {
		go func() {
			for {
				time.Sleep(1 * time.Second)
				log.Printf("rooms: %d, users: %d, clients: %d \n", goio.Rooms().Count(), goio.Users().Count(), goio.Clients().Count())
			}
		}()
	}

	goio.LifeCycle = *flagClientLifeCycle

	m := gin.New()
	m.Use(gin.Recovery())
	if *flagDebug {
		m.Use(gin.Logger())
	}

	m.Use(func(ctx *gin.Context) {
		ctx.Header("Content-Type", "text/plain; charset=utf-8")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Allow-Methods", "GET,POST")
		ctx.Header("Powered-By", "Goio")

		if "" != *flagAllowOrigin {
			allowOrigins := strings.Split(*flagAllowOrigin, ",")
			for _, allowOrigin := range allowOrigins {
				ctx.Writer.Header().Add("Access-Control-Allow-Origin", allowOrigin)
			}
		}
	})

	if *flagDebug {

		m.GET("/test/message", func(ctx *gin.Context) {
			userId1 := "u1"
			user1 := goio.Users().MustGet(userId1)
			clt1 := goio.NewClient(user1)

			roomId := "r1"
			room := goio.Rooms().Get(roomId)
			if room == nil {
				room = goio.NewRoom(roomId)
			}

			goio.AddRoomUser(room, user1)

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

			ctx.String(200, "completed")
		})

		m.GET("/test/client/:num", func(ctx *gin.Context) {

			count, _ := strconv.Atoi(ctx.Param("num"))
			fmt.Printf("adding %d clients\n", count)

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

					goio.AddRoomUser(room, user)

					all <- struct{}{}
				}(i)
			}

			for i := 0; i < count; i++ {
				<-all
			}

			ctx.String(200, fmt.Sprintf("%d s", time.Now().Second()-st))
		})
	}

	m.GET("/count", func(ctx *gin.Context) {
		ctx.String(200, fmt.Sprintf("rooms: %d, users: %d, clients: %d \n", goio.Rooms().Count(), goio.Users().Count(), goio.Clients().Count()))
	})

	// get user ids array from the given room
	m.GET("/room/users/:room_id", func(ctx *gin.Context) {
		roomId := ctx.Param("room_id")

		room := goio.Rooms().Get(roomId)
		if room == nil {
			ctx.String(200, "")
			return
		}

		ctx.String(200, strings.Join(room.UserIds(), ","))
	})

	m.GET("/user/data/:user_id/:key", func(ctx *gin.Context) {
		userId := ctx.Param("user_id")
		key := ctx.Param("key")

		user := goio.Users().Get(userId)
		if user == nil {
			ctx.String(200, "")
			return
		}

		ctx.String(200, user.GetData(key))
	})

	m.POST("/user/data/:client_id/:key", func(ctx *gin.Context) {
		val, err := ioutil.ReadAll(ctx.Request.Body)
		ctx.Request.Body.Close()

		if err != nil {
			ctx.String(500, err.Error())
			return
		}

		clientId := ctx.Param("client_id")
		key := ctx.Param("key")

		clt := goio.Clients().Get(clientId)
		if clt != nil && clt.User != nil {
			clt.User.AddData <- goio.UserData{Key: key, Val: string(val)}
		}

		ctx.String(204, "")
	})

	m.POST("/online_status", func(ctx *gin.Context) {
		val, err := ioutil.ReadAll(ctx.Request.Body)
		ctx.Request.Body.Close()

		if err != nil {
			fmt.Printf("online_status error %s\n", err.Error())
			ctx.String(500, err.Error())
			return
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

		ctx.String(200, status)
	})

	m.POST("/client/:user_id", func(ctx *gin.Context) {
		userId := ctx.Param("user_id")
		user := goio.Users().MustGet(userId)
		clt := goio.NewClient(user)

		ctx.String(200, clt.Id)
	})

	m.GET("/kill_client/:client_id", func(ctx *gin.Context) {
		clt := goio.Clients().Get(ctx.Param("client_id"))
		if clt != nil {
			clt.User.DelClt <- clt
		}

		ctx.String(204, "")
	})

	m.GET("/message/:client_id", func(ctx *gin.Context) {
		id := ctx.Param("client_id")
		clt := goio.Clients().Get(id)

		if clt == nil {
			ctx.String(404, fmt.Sprintf("Client %s does not exist\n", id))
			return
		}

		for i, wait := 0, 1; i*wait <= *flagClientMessageTimeout; i++ {

			// keep this client alive
			clt.Handshake()

			msgs := clt.Msgs()
			if len(msgs) == 0 {
				time.Sleep(time.Duration(wait) * time.Second)
				continue
			}

			m, err := json.Marshal(msgs)
			if err != nil {
				ctx.String(500, err.Error())
				return
			}

			ctx.String(200, string(m))
			return
		}

		ctx.String(200, "")
	})

	m.POST("/message/:client_id", func(ctx *gin.Context) {
		id := ctx.Param("client_id")
		clt := goio.Clients().Get(id)
		if clt == nil {
			ctx.String(403, fmt.Sprintf("Client %s does not exist\n", id))
			return
		}

		if clt.User == nil {
			ctx.String(403, fmt.Sprintf("Client %s does not connect with any user\n", id))
			return
		}

		msg := &goio.Message{}
		ctx.BindJSON(msg)

		msg.CallerId = clt.User.Id
		msg.ClientId = clt.Id

		clt.Handshake()
		// we handshake again after it finished no matter timeout or ok
		defer clt.Handshake()

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

					goio.AddRoomUser(r, clt.User)
					r.Message <- msg
				}

			case "leave":
				fmt.Printf("msg event - leave\n")
				if msg.RoomId != "" {
					r := goio.Rooms().Get(msg.RoomId)
					if r != nil {
						goio.DelRoomUser(r, clt.User)
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

		ctx.String(204, "")
	})

	m.OPTIONS("/*path", func(ctx *gin.Context) {})

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
