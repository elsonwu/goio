package goio

import (
	"log"
	"time"

	"github.com/golang/glog"
)

var gcIsRunning = false

func gc() {

	if gcIsRunning {
		return
	}

	startTime := time.Now()
	gcIsRunning = true
	defer func() {
		log.Println("GC process " + time.Now().Sub(startTime).String())
		gcIsRunning = false
	}()

	var clt *Client
	Clients().m.Range(func(k interface{}, v interface{}) bool {
		clt = v.(*Client)
		if clt == nil || !clt.IsDead() {
			return true
		}

		glog.V(1).Infoln("clt " + clt.Id + " is dead")
		Clients().DelClt(clt.Id)
		clt.User.DelClt(clt.Id)

		return true
	})

	var deadUsrs []*User
	Users().m.Range(func(k interface{}, v interface{}) bool {
		u := v.(*User)
		if u == nil || !u.IsDead() {
			return true
		}

		deadUsrs = append(deadUsrs, u)

		glog.V(1).Infoln("user " + u.Id + " is dead")
		Users().delUser(u.Id)
		for _, r := range u.Rooms() {
			r.delUser(u.Id)
		}

		return true
	})

	var r *Room
	Rooms().m.Range(func(k interface{}, v interface{}) bool {
		r = v.(*Room)
		if r == nil || !r.IsDead() {
			return true
		}

		glog.V(1).Infoln("room " + r.Id + " is dead")
		Rooms().DelRoom(r.Id)
		return true
	})

	// tell everyone these users are offline
	for _, u := range deadUsrs {
		if !u.IsDead() {
			continue
		}

		SendMessage(&Message{
			EventName: MsgLeave,
			CallerId:  u.Id,
		}, u)
	}
}
