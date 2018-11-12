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

type EventManager struct {
	fields       map[string][]EventHandler
	errorHandler ErrorHandler
	preHandlers  []EventHandler
	postHandlers []EventHandler
}

func NewEventManager() *EventManager {
	return &EventManager{
		fields:       make(map[string][]EventHandler),
		errorHandler: func(ctx context.Context, event *Event, err error) {},
		preHandlers:  []EventHandler{},
		postHandlers: []EventHandler{},
	}
}

func (e *EventManager) OnError(handler ErrorHandler) {
	e.errorHandler = handler
}

func (e *EventManager) RegisterField(field string, handlers ...EventHandler) {
	e.fields[field] = handlers
}

func (e *EventManager) RegisterPreFunction(handlers ...EventHandler) {
	if len(handlers) == 0 {
		return
	}

	e.preHandlers = handlers
}

func (e *EventManager) RegisterPostFunction(handlers ...EventHandler) {
	if len(handlers) == 0 {
		return
	}

	e.postHandlers = handlers
}

func (e *EventManager) runHandleError(ctx context.Context, event *Event, err error, data interface{}) (*Result, error) {
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

func (e *EventManager) Run(ctx context.Context, event *Event) (*Result, error) {
	mainHandlers, ok := e.fields[event.Field]
	if ok && len(mainHandlers) > 0 {
		var data interface{}
		var err error
		for _, preHandler := range e.preHandlers {
			if _, err = preHandler(ctx, event); err != nil {
				return e.runHandleError(ctx, event, err, data)
			}
		}

		for _, handler := range mainHandlers {
			result, err := handler(ctx, event)
			if err != nil {
				return e.runHandleError(ctx, event, err, data)
			}

			if result != nil {
				data = result
			}
		}

		for _, postHandler := range e.postHandlers {
			if _, err = postHandler(ctx, event); err != nil {
				return e.runHandleError(ctx, event, err, data)
			}
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
