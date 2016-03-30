package goio

func DelRoomUser(room *Room, user *User) {
	room.delUser <- user
	user.delRoom <- room
}

func AddRoomUser(room *Room, user *User) {
	room.addUser <- user
	user.addRoom <- room
}
