package goio

type Message struct {
	EventName string `json:"e"` // join/leave/broadcast/connect
	Data      string `json:"d"`
	RoomId    string `json:"r"`
	CallerId  string `json:"c"`
	ClientId  string `json:"-"` // don't want to output this field
}
