package goio

func DelRoomUser(room *Room, user *User) {
	room.delUser <- user
	user.delRoom <- room
}

func AddRoomUser(room *Room, user *User) {
	room.addUser <- user
	user.addRoom <- room
}

func AddUserClt(user *User, clt *Client) {
	Clients().addClt <- clt
	user.addClt <- clt
}

func DelUserClt(user *User, clt *Client) {
	Clients().delClt <- clt
	user.delClt <- clt
}
