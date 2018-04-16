package goio

import (
	"reflect"
	"time"

	"github.com/golang/glog"
)

const (
	MsgJoin      = `join`
	MsgLeave     = `leave`
	MsgBroadcast = `broadcast`
)

func Run() {
	glog.V(1).Infoln("init users/rooms/clients")
	Users()
	Rooms()
	Clients()

	glog.V(1).Infof("gc file - gc period:%d\n", GCPeriod)
	go func() {
		for {
			<-time.After(time.Duration(GCPeriod) * time.Second)
			glog.V(1).Infoln("Running GC")
			gc()
		}
	}()
}

func SendMessage(msg *Message, caller interface{}) {

	glog.V(1).Infoln("SendMessage " + msg.EventName + " caller " + reflect.TypeOf(caller).String())

	switch msg.EventName {
	case MsgJoin:
		if msg.RoomId != "" {
			clt, ok := caller.(*Client)
			if !ok || clt.IsDead() || clt.User.IsDead() {
				return
			}

			r := Rooms().MustGet(msg.RoomId)
			r.addUser(clt.User)
			clt.User.AddRoom(r)

			go r.addMessage(msg)
		} else {
			// tell everyone this new client online
			go Clients().addMessage(msg)
		}

	case MsgLeave:
		// tell all users in the same room the caller dropout
		if msg.RoomId != "" {
			r := Rooms().Get(msg.RoomId)
			if r != nil {
				clt, ok := caller.(*Client)
				if !ok {
					return
				}

				r.delUser(clt.User)
				clt.User.DelRoom(r)

				go r.addMessage(msg)
			}
		} else if msg.CallerId != "" {
			// tell everyone this client offline
			go Clients().addMessage(msg)
		}

	case MsgBroadcast:
		if msg.RoomId != "" {
			r := Rooms().Get(msg.RoomId)
			if r != nil {
				go r.addMessage(msg)
			}
		} else {
			clt, ok := caller.(*Client)
			if !ok {
				return
			}

			for _, r := range clt.User.Rooms() {
				go r.addMessage(msg)
			}
		}
	}
}
