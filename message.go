package goreal

type Message struct {
	EventName string      `json:"event_name"`
	Data      interface{} `json:"data"`
}
