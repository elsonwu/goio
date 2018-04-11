package goio

import (
	"github.com/golang/glog"
)

func gc() {
	var clt *Client
	Clients().m.Range(func(k interface{}, v interface{}) bool {
		clt = v.(*Client)
		if clt == nil || !clt.IsDead() {
			return true
		}

		glog.V(1).Infoln("clt " + clt.Id + " is dead")
		Clients().DelClt(clt)
		clt.User.DelClt(clt)

		return true
	})

	var deadUsrs []*User
	var u *User
	Users().m.Range(func(k interface{}, v interface{}) bool {
		u = v.(*User)
		if u == nil || !u.IsDead() {
			return true
		}

		deadUsrs = append(deadUsrs, u)

		glog.V(1).Infoln("user " + u.Id + " is dead")
		Users().delUser(u)
		for _, r := range u.Rooms() {
			r.delUser(u)
		}

		return true
	})

	var r *Room
	Rooms().m.Range(func(k interface{}, v interface{}) bool {
		r = v.(*Room)
		if r == nil || !r.IsDead() {
			return true
		}

		glog.V(1).Infoln("room " + u.Id + " is dead")
		Rooms().DelRoom(r)
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
