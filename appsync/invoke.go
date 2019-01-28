package appsync

import (
	"context"
	"encoding/json"
)

type InvokePreHandler func(ctx context.Context, event *InvokeEvent) error
type InvokePostHandler func(ctx context.Context, event *InvokeEvent, result interface{}, err error)
type InvokeEventHandler func(ctx context.Context, event *InvokeEvent) (interface{}, error)
type InvokeErrorHandler func(ctx context.Context, event *InvokeEvent, err error)

type invokeHandlers struct {
	handler      InvokeEventHandler
	preHandlers  []InvokePreHandler
	postHandlers []InvokePostHandler
}

type InvokeEvents []*InvokeEvent

type InvokeEvent struct {
	Field    string          `json:"field"`
	Args     json.RawMessage `json:"arguments"`
	Source   json.RawMessage `json:"source"`
	Identity *Identity       `json:"identity"`
}

func (e *InvokeEvent) ParseArgs(v interface{}) error {
	return json.Unmarshal(e.Args, v)
}

func (e *InvokeEvent) ParseSource(v interface{}) error {
	return json.Unmarshal(e.Source, v)
}
