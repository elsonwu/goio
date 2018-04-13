package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/elsonwu/goio"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

var flagHost = flag.String("host", "127.0.0.1:9999", "the default server host")
var flagSSLHost = flag.String("sslhost", "", "the server host for https, it will override the host setting")
var flagAllowOrigin = flag.String("alloworigin", "*", "the host allow to cross site ajax")
var flagDebug = flag.Bool("debug", false, "enable debug mode or not")
var flagGCPeriod = flag.Int("gcperiod", 5, "how many seconds to run gc")
var flagClientLifeCycle = flag.Int64("lifecycle", 15, "how many seconds of the client life cycle")
var flagClientMessageTimeout = flag.Int("messagetimeout", 15, "how many seconds of the client keep waiting for new messages")
var flagEnableHttps = flag.Bool("enablehttps", false, "enable https or not")
var flagDisableHttp = flag.Bool("disablehttp", false, "disable http and use https only")
var flagCertFile = flag.String("certfile", "", "certificate file path")
var flagKeyFile = flag.String("keyfile", "", "private file path")

func main() {

	flag.Parse()

	goio.GCPeriod = *flagGCPeriod
	goio.LifeCycle = *flagClientLifeCycle
	goio.Run()

	m := gin.New()
	m.Use(gin.Recovery())
	if *flagDebug {
		m.Use(gin.Logger())
	} else {
		gin.SetMode(gin.ReleaseMode)
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

	m.Use(func(ctx *gin.Context) {
		ctx.Next()
		ctx.Request.Close = true
		ctx.Request.Body.Close()
	})

	m.GET("/count", func(ctx *gin.Context) {

		s := fmt.Sprintf("rooms: %d, users: %d, clients: %d \n", goio.Rooms().Count(), goio.Users().Count(), goio.Clients().Count())
		if yes, _ := ctx.GetQuery("debug"); yes == "1" {
			s = s + "#Rooms\n"
			goio.Rooms().Range(func(r *goio.Room) {
				s = s + " - " + r.Id + " user count:" + strconv.Itoa(r.UserCount()) + "\n"
			})

			s = s + "\n"

			s = s + "#Users\n"
			goio.Users().Range(func(u *goio.User) {
				s = s + " - " + u.Id + " client count:" + strconv.Itoa(u.ClientCount()) + "\n"
			})

			s = s + "\n"

			s = s + "#Clients\n"
			goio.Clients().Range(func(c *goio.Client) {
				s = s + " - " + c.Id + " user ID:" + c.User.Id + " user client count:" + strconv.Itoa(c.User.ClientCount()) + "\n"
			})
		}

		ctx.String(200, s)
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
			clt.User.AddData(key, string(val))
		}

		ctx.String(204, "")
	})

	m.POST("/online_status", func(ctx *gin.Context) {
		val, err := ioutil.ReadAll(ctx.Request.Body)
		ctx.Request.Body.Close()

		if err != nil {
			glog.V(1).Infof("online_status error %s\n", err.Error())
			ctx.String(500, err.Error())
			return
		}

		userIds := strings.Split(string(val), ",")
		glog.V(1).Infof("checking userIds:%s\n", string(val))

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
			clt.SetIsDead()
		}

		ctx.String(204, "")
	})

	m.GET("/message/:client_id", func(ctx *gin.Context) {
		id := ctx.Param("client_id")
		clt := goio.Clients().Get(id)

		if clt == nil || clt.IsDead() {
			ctx.String(404, fmt.Sprintf("Client %s does not exist\n", id))
			return
		}

		for i := 0; i <= *flagClientMessageTimeout; i++ {
			msgs := clt.ReadMessages()
			if len(msgs) == 0 {
				<-time.After(time.Second)
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
		if clt == nil || clt.IsDead() {
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

		goio.SendMessage(msg, clt)

		ctx.String(204, "")
	})

	m.OPTIONS("/*path", func(ctx *gin.Context) {})

	// Register pprof handlers
	ginpprof.Wrap(m)

	host := *flagHost
	if !*flagEnableHttps && *flagDisableHttp {
		glog.Errorln("You cannot disable http but not enable https in the same time")
		return
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
