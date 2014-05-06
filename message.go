package goio

type Message struct {
	EventName string `json:"e"`
	Data      string `json:"d"`
	RoomId    string `json:"r"`
	CallerId  string `json:"c"`
}
