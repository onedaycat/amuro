package appsync

import (
	"context"
	"encoding/json"
)

type BatchInvokePreHandler func(ctx context.Context, event *BatchInvokeEvent) error
type BatchInvokePostHandler func(ctx context.Context, event *BatchInvokeEvent, result interface{}, err error)
type BatchInvokeEventHandler func(ctx context.Context, event *BatchInvokeEvent) (interface{}, error)
type BatchInvokeErrorHandler func(ctx context.Context, event *BatchInvokeEvent, err error)

type batchInvokeHandlers struct {
	handler      BatchInvokeEventHandler
	preHandlers  []BatchInvokePreHandler
	postHandlers []BatchInvokePostHandler
}

type BatchInvokeEvent struct {
	Field    string          `json:"field"`
	Args     json.RawMessage `json:"arguments"`
	Sources  json.RawMessage `json:"sources"`
	Identity *Identity       `json:"identity"`
	NSource  int             `json:"-"`
}

func (e *BatchInvokeEvent) ParseArgs(v interface{}) error {
	return json.Unmarshal(e.Args, v)
}

func (e *BatchInvokeEvent) ParseSources(v interface{}) error {
	return json.Unmarshal(e.Sources, v)
}
