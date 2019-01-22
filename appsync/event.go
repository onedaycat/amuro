package appsync

import (
	"context"
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/onedaycat/errors"
)

type PreHandler func(ctx context.Context, event *Event) error
type PostHandler func(ctx context.Context, event *Event, result interface{}, err error)
type EventHandler func(ctx context.Context, event *Event) (interface{}, error)
type ErrorHandler func(ctx context.Context, event *Event, err error)

type Event struct {
	Field       string          `json:"field"`
	Args        json.RawMessage `json:"arguments"`
	Source      json.RawMessage `json:"source"`
	Identity    *Identity       `json:"identity"`
	BatchSource []map[string]interface{}
}

func (e *Event) UnmarshalJSON(b []byte) error {
	valRoot, dataTypeRoot, _, err := jsonparser.Get(b)
	if err != nil {
		return err
	}

	if dataTypeRoot == jsonparser.Array {
		e.BatchSource = make([]map[string]interface{}, 0, 5)
		index := 0
		jsonparser.ArrayEach(valRoot, func(valRootArr []byte, dataTypeRootArr jsonparser.ValueType, offset int, err error) {
			var sourceVal []byte
			var dataTypeSource jsonparser.ValueType
			var idVal []byte
			var dataTypeID jsonparser.ValueType
			if index == 0 {
				if e.Field, err = jsonparser.GetString(valRootArr, "field"); err != nil && err != jsonparser.KeyPathNotFoundError {
					panic(err)
				}

				if idVal, dataTypeID, _, err = jsonparser.Get(valRootArr, "identity"); err != nil && err != jsonparser.KeyPathNotFoundError {
					panic(err)
				}
				if dataTypeID == jsonparser.Object {
					e.Identity = &Identity{}
					json.Unmarshal(idVal, e.Identity)
				}

				if e.Args, _, _, err = jsonparser.Get(valRootArr, "arguments"); err != nil && err != jsonparser.KeyPathNotFoundError {
					panic(err)
				}

			}

			if sourceVal, dataTypeSource, _, err = jsonparser.Get(valRootArr, "source"); err != nil && err != jsonparser.KeyPathNotFoundError {
				panic(err)
			}

			if dataTypeSource == jsonparser.Object {
				source := make(map[string]interface{})
				if err = json.Unmarshal(sourceVal, &source); err != nil {
					panic(err)
				}
				e.BatchSource = append(e.BatchSource, source)
			}

			index++
		})

		return nil
	} else if dataTypeRoot == jsonparser.Object {
		var err error
		var idVal []byte
		var dataTypeID jsonparser.ValueType
		if e.Field, err = jsonparser.GetString(valRoot, "field"); err != nil && err != jsonparser.KeyPathNotFoundError {
			panic(err)
		}

		if idVal, dataTypeID, _, err = jsonparser.Get(valRoot, "identity"); err != nil && err != jsonparser.KeyPathNotFoundError {
			panic(err)
		}
		if dataTypeID == jsonparser.Object {
			e.Identity = &Identity{}
			json.Unmarshal(idVal, e.Identity)
		}

		if e.Args, _, _, err = jsonparser.Get(valRoot, "arguments"); err != nil && err != jsonparser.KeyPathNotFoundError {
			panic(err)
		}

		if e.Source, _, _, err = jsonparser.Get(valRoot, "source"); err != nil && err != jsonparser.KeyPathNotFoundError {
			panic(err)
		}

		return err
	}

	return errors.Newf("Unable to UnmarshalJSON of %s", dataTypeRoot.String())
}

type Result struct {
	Data  interface{}      `json:"data"`
	Error *errors.AppError `json:"error"`
}

func (e *Event) ParseArgs(v interface{}) error {
	return json.Unmarshal(e.Args, v)
}

func (e *Event) ParseSource(v interface{}) error {
	return json.Unmarshal(e.Source, v)
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
