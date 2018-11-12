package appsync

import (
	"context"
	"encoding/json"

	"github.com/onedaycat/errors"
)

type PreHandler func(ctx context.Context, event *Event) error
type PostHandler func(ctx context.Context, event *Event, result interface{}, err error)
type EventHandler func(ctx context.Context, event *Event) (interface{}, error)
type ErrorHandler func(ctx context.Context, event *Event, err error)

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

type MainHandler struct {
	handler      EventHandler
	preHandlers  []PreHandler
	postHandlers []PostHandler
}

type EventManager struct {
	fields       map[string]*MainHandler
	errorHandler ErrorHandler
	preHandlers  []PreHandler
	postHandlers []PostHandler
}

func NewEventManager() *EventManager {
	return &EventManager{
		fields:       make(map[string]*MainHandler),
		errorHandler: func(ctx context.Context, event *Event, err error) {},
		preHandlers:  []PreHandler{},
		postHandlers: []PostHandler{},
	}
}

func (e *EventManager) OnError(handler ErrorHandler) {
	e.errorHandler = handler
}

func (e *EventManager) RegisterField(field string, handler EventHandler, preHandler []PreHandler, postHandler []PostHandler) {
	e.fields[field] = &MainHandler{
		handler:      handler,
		preHandlers:  preHandler,
		postHandlers: postHandler,
	}
}

func (e *EventManager) UsePreHandler(handlers ...PreHandler) {
	if len(handlers) == 0 {
		return
	}

	e.preHandlers = handlers
}

func (e *EventManager) UsePostHandler(handlers ...PostHandler) {
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

func (e *EventManager) runPreHandler(ctx context.Context, event *Event, handlers []PreHandler) error {
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (e *EventManager) runPostHandler(ctx context.Context, event *Event, data interface{}, handlerErr error, handlers []PostHandler) {
	for _, handler := range handlers {
		handler(ctx, event, data, handlerErr)
	}
}

func (e *EventManager) Run(ctx context.Context, event *Event) (*Result, error) {
	if mainHandler, ok := e.fields[event.Field]; ok {
		if err := e.runPreHandler(ctx, event, e.preHandlers); err != nil {
			return e.runHandleError(ctx, event, err, nil)
		}

		if err := e.runPreHandler(ctx, event, mainHandler.preHandlers); err != nil {
			return e.runHandleError(ctx, event, err, nil)
		}

		data, err := mainHandler.handler(ctx, event)

		e.runPostHandler(ctx, event, data, err, mainHandler.postHandlers)
		e.runPostHandler(ctx, event, data, err, e.postHandlers)

		if err != nil {
			return e.runHandleError(ctx, event, err, data)
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
