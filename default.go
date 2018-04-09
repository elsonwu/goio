package goio

import "sync"

// for client life cycle, default 30s
var LifeCycle int64 = 30

var _us *users
var _rs *rooms
var _cs *clients

var usLock sync.Mutex
var rsLock sync.Mutex
var csLock sync.Mutex

func init() {
	Users()
	Rooms()
	Clients()
}

func Users() *users {
	if _us == nil {
		usLock.Lock()

		if _us == nil {
			_us = NewUsers()
		}

		usLock.Unlock()
	}

	return _us
}

func Rooms() *rooms {
	if _rs == nil {
		rsLock.Lock()

		if _rs == nil {
			_rs = NewRooms()
		}

		rsLock.Unlock()
	}

	return _rs
}

func Clients() *clients {
	if _cs == nil {
		csLock.Lock()

		if _cs == nil {
			_cs = NewClients()
		}

		csLock.Unlock()
	}

	return _cs
}
