package goreal

type Message struct {
	EventName string      `json:"n"`
	Data      interface{} `json:"d"`
	RoomId    string      `json:"r"`
	CallerId  string      `json:"c"`
}
