package goio

import "time"

func DelRoomUser(room *Room, user *User) {
	select {
	case room.delUser <- user:
	case <-time.After(10 * time.Second):
	}

	select {
	case user.delRoom <- room:
	case <-time.After(10 * time.Second):
	}
}

func AddRoomUser(room *Room, user *User) {

	select {
	case room.addUser <- user:
	case <-time.After(10 * time.Second):
	}

	select {
	case user.addRoom <- room:
	case <-time.After(10 * time.Second):
	}
}

func AddUserClt(user *User, clt *Client) {
	Clients().addClt <- clt

	select {
	case user.addClt <- clt:
	case <-time.After(10 * time.Second):
	}
}

func DelUserClt(user *User, clt *Client) {
	Clients().delClt <- clt

	select {
	case user.delClt <- clt:
	case <-time.After(10 * time.Second):
	}
}
