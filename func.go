package goio

func DelRoomUser(room *Room, user *User) {
	room.DelUser(user)
	user.DelRoom(room)
}

func AddRoomUser(room *Room, user *User) {
	room.AddUser(user)
	user.AddRoom(room)
}

func AddUserClt(user *User, clt *Client) {
	Clients().AddClt(clt)
	user.AddClt(clt)
}

func DelUserClt(user *User, clt *Client) {
	Clients().DelClt(clt)
	user.DelClt(clt)
}
