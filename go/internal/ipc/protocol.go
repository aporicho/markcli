package ipc

import "encoding/json"

type Response struct {
	Type    string          `json:"type"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type Request struct {
	ID     string
	Method string
	Params json.RawMessage
	reply  chan Response
}

func NewRequest(id, method string, params json.RawMessage) Request {
	return Request{ID: id, Method: method, Params: params, reply: make(chan Response, 1)}
}

func (r Request) Reply(resp Response) { r.reply <- resp }
func (r Request) Chan() <-chan Response { return r.reply }
