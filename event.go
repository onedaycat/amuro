package appsync

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/onedaycat/errors"
)

type Identity struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	UserArn  string `json:"userArn"`
}

type Event struct {
	Field    string          `json:"field"`
	Args     json.RawMessage `json:"args"`
	Source   json.RawMessage `json:"source"`
	Identity *Identity       `json:"identity"`
}

type Result struct {
	Data  interface{}      `json:"data"`
	Error *errors.AppError `json:"error"`
}

func (e *Event) ParseArgs(v interface{}) error {
	return json.Unmarshal(e.Args, v)
}

func (e *Event) ParseSource(v interface{}) error {
	return json.Unmarshal(e.Args, v)
}

type EventHandler func(ctx context.Context, event *Event) (interface{}, error)

type EventManager struct {
	fields map[string]EventHandler
}

func NewEventManager() *EventManager {
	return &EventManager{
		fields: make(map[string]EventHandler),
	}
}

func (e *EventManager) RegisterField(field string, handler EventHandler) {
	e.fields[field] = handler
}

func (e *EventManager) Run(ctx context.Context, event *Event) (*Result, error) {
	if handler, ok := e.fields[event.Field]; ok {
		data, err := handler(ctx, event)
		if err != nil {
			appErr, ok := errors.FromError(err)
			if ok {
				return &Result{
					Data:  data,
					Error: appErr,
				}, nil
			}

			return &Result{
				Data:  data,
				Error: errors.InternalError("UNKNOWN_CODE", err.Error()),
			}, nil
		}

		return &Result{
			Data:  data,
			Error: nil,
		}, nil
	}

	return nil, fmt.Errorf("Not found handler on field %s", event.Field)
}
