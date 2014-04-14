package goreal

import (
	"strconv"
	"time"
)

func NewClientRoomHandler() ClientRoomHandler {
	return make(ClientRoomHandler)
}

func NewClientHandler() ClientHandler {
	return make(ClientHandler)
}

func uuid() string {
	return strconv.Itoa(int(time.Now().Nanosecond()))
}
