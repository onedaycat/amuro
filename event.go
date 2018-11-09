package appsync

import (
	"context"
	"encoding/json"

	"github.com/onedaycat/errors"
)

type Identity struct {
	ID     string   `json:"id"`
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
	IP     string   `json:"ip"`
}

type Event struct {
	Field    string          `json:"field"`
	Args     json.RawMessage `json:"arguments"`
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
type ErrorHandler func(ctx context.Context, event *Event, err error)
type Middleware func(next EventHandler) EventHandler

type EventManager struct {
	fields       map[string]EventHandler
	errorHandler ErrorHandler
	middlewares  []EventHandler
}

func NewEventManager() *EventManager {
	return &EventManager{
		fields:       make(map[string]EventHandler),
		errorHandler: func(ctx context.Context, event *Event, err error) {},
	}
}

func (e *EventManager) OnError(handler ErrorHandler) {
	e.errorHandler = handler
}

func (e *EventManager) RegisterField(field string, handler EventHandler) {
	e.fields[field] = handler
}

func (e *EventManager) RegisterMiddleware(middleware EventHandler) {
	e.middlewares = append(e.middlewares, middleware)
}

func (e *EventManager) Run(ctx context.Context, event *Event) (interface{}, error) {
	for _, middlewareFunc := range e.middlewares {
		_, err := middlewareFunc(ctx, event)
		if err != nil {
			return nil, err
		}
	}

	return e.runHandler(ctx, event)
}

func (e *EventManager) runHandler(ctx context.Context, event *Event) (*Result, error) {
	if handler, ok := e.fields[event.Field]; ok {
		data, err := handler(ctx, event)
		if err != nil {
			e.errorHandler(ctx, event, err)
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

	err := errors.InternalErrorf("FIELD_NOT_FOUND", "Not found handler on field %s", event.Field)
	e.errorHandler(ctx, event, err)

	return nil, err
}

func (e *EventManager) DefaultHandler(ctx context.Context, event *Event) (interface{}, error) {
	return e.Run(ctx, event)
}
