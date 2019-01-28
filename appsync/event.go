package appsync

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/buger/jsonparser"
	"github.com/onedaycat/errors"
)

const (
	eventInvokeType      int = 0
	eventBatchInvokeType int = 1
)

type request struct {
	InvokeEvent      *InvokeEvent
	BatchInvokeEvent *BatchInvokeEvent
	InvokeEvents     []*InvokeEvent
	eventType        int
	sources          json.RawMessage
}

func (r *request) UnmarshalJSON(b []byte) error {
	_, dataTypeRoot, _, err := jsonparser.Get(b)
	if err != nil {
		return err
	}

	if dataTypeRoot == jsonparser.Array {
		r.InvokeEvents = make(InvokeEvents, 0, 5)
		r.eventType = eventBatchInvokeType
		if err = json.Unmarshal(b, &r.InvokeEvents); err != nil {
			return err
		}

		if len(r.InvokeEvents) == 0 {
			return errors.Newf("No data in batch invoke")
		}

		b := bytes.NewBuffer(nil)
		b.WriteByte(91)
		first := true
		for i := 0; i < len(r.InvokeEvents); i++ {
			if len(r.InvokeEvents[i].Source) == 0 {
				continue
			}

			if !first {
				b.WriteByte(44)
			}
			b.Write(r.InvokeEvents[i].Source)
			first = false
		}
		b.WriteByte(93)

		r.BatchInvokeEvent = &BatchInvokeEvent{
			Field:    r.InvokeEvents[0].Field,
			Args:     r.InvokeEvents[0].Args,
			Sources:  b.Bytes(),
			Identity: r.InvokeEvents[0].Identity,
		}

		if len(r.BatchInvokeEvent.Sources) == 2 {
			r.BatchInvokeEvent.Sources = nil
		}

		return nil
	} else if dataTypeRoot == jsonparser.Object {
		r.InvokeEvent = &InvokeEvent{}
		r.eventType = eventInvokeType
		return json.Unmarshal(b, r.InvokeEvent)
	}

	return errors.Newf("Unable to UnmarshalJSON of %s", dataTypeRoot.String())
}

type Result struct {
	Data  interface{}      `json:"data"`
	Error *errors.AppError `json:"error"`
}

type EventManager struct {
	invokeFields            map[string]*invokeHandlers
	invokeErrorHandler      InvokeErrorHandler
	invokePreHandlers       []InvokePreHandler
	invokePostHandlers      []InvokePostHandler
	batchInvokeFields       map[string]*batchInvokeHandlers
	batchInvokeErrorHandler BatchInvokeErrorHandler
	batchInvokePreHandlers  []BatchInvokePreHandler
	batchInvokePostHandlers []BatchInvokePostHandler
}

func NewEventManager() *EventManager {
	return &EventManager{
		invokeFields:            make(map[string]*invokeHandlers),
		invokeErrorHandler:      func(ctx context.Context, event *InvokeEvent, err error) {},
		invokePreHandlers:       []InvokePreHandler{},
		invokePostHandlers:      []InvokePostHandler{},
		batchInvokeFields:       make(map[string]*batchInvokeHandlers),
		batchInvokeErrorHandler: func(ctx context.Context, event *BatchInvokeEvent, err error) {},
		batchInvokePreHandlers:  []BatchInvokePreHandler{},
		batchInvokePostHandlers: []BatchInvokePostHandler{},
	}
}

func (e *EventManager) OnInvokeError(handler InvokeErrorHandler) {
	e.invokeErrorHandler = handler
}

func (e *EventManager) OnBatchInvokeError(handler BatchInvokeErrorHandler) {
	e.batchInvokeErrorHandler = handler
}

func (e *EventManager) RegisterInvoke(field string, handler InvokeEventHandler, preHandler []InvokePreHandler, postHandler []InvokePostHandler) {
	e.invokeFields[field] = &invokeHandlers{
		handler:      handler,
		preHandlers:  preHandler,
		postHandlers: postHandler,
	}
}

func (e *EventManager) RegisterBatchInvoke(field string, handler BatchInvokeEventHandler, preHandler []BatchInvokePreHandler, postHandler []BatchInvokePostHandler) {
	e.batchInvokeFields[field] = &batchInvokeHandlers{
		handler:      handler,
		preHandlers:  preHandler,
		postHandlers: postHandler,
	}
}

func (e *EventManager) UseInvokePreHandler(handlers ...InvokePreHandler) {
	if len(handlers) == 0 {
		return
	}

	e.invokePreHandlers = handlers
}

func (e *EventManager) UseBatchInvokePreHandler(handlers ...BatchInvokePreHandler) {
	if len(handlers) == 0 {
		return
	}

	e.batchInvokePreHandlers = handlers
}

func (e *EventManager) UseInvokePostHandler(handlers ...InvokePostHandler) {
	if len(handlers) == 0 {
		return
	}

	e.invokePostHandlers = handlers
}

func (e *EventManager) UseBatchInvokePostHandler(handlers ...BatchInvokePostHandler) {
	if len(handlers) == 0 {
		return
	}

	e.batchInvokePostHandlers = handlers
}

func (e *EventManager) runHandleInvokeError(ctx context.Context, event *InvokeEvent, err error, data interface{}) (*Result, error) {
	e.invokeErrorHandler(ctx, event, err)
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

func (e *EventManager) runInvokePreHandler(ctx context.Context, event *InvokeEvent, handlers []InvokePreHandler) error {
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (e *EventManager) runInvokePostHandler(ctx context.Context, event *InvokeEvent, data interface{}, handlerErr error, handlers []InvokePostHandler) {
	for _, handler := range handlers {
		handler(ctx, event, data, handlerErr)
	}
}

func (e *EventManager) runHandleBatchInvokeError(ctx context.Context, event *BatchInvokeEvent, err error, data interface{}) (*Result, error) {
	e.batchInvokeErrorHandler(ctx, event, err)
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

func (e *EventManager) runBatchInvokePreHandler(ctx context.Context, event *BatchInvokeEvent, handlers []BatchInvokePreHandler) error {
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (e *EventManager) runBatchInvokePostHandler(ctx context.Context, event *BatchInvokeEvent, data interface{}, handlerErr error, handlers []BatchInvokePostHandler) {
	for _, handler := range handlers {
		handler(ctx, event, data, handlerErr)
	}
}

func (e *EventManager) Run(ctx context.Context, req *request) (*Result, error) {
	switch req.eventType {
	case eventBatchInvokeType:
		event := req.BatchInvokeEvent
		if mainHandler, ok := e.batchInvokeFields[event.Field]; ok {
			if err := e.runBatchInvokePreHandler(ctx, event, e.batchInvokePreHandlers); err != nil {
				return e.runHandleBatchInvokeError(ctx, event, err, nil)
			}

			if err := e.runBatchInvokePreHandler(ctx, event, mainHandler.preHandlers); err != nil {
				return e.runHandleBatchInvokeError(ctx, event, err, nil)
			}

			data, err := mainHandler.handler(ctx, event)

			e.runBatchInvokePostHandler(ctx, event, data, err, mainHandler.postHandlers)
			e.runBatchInvokePostHandler(ctx, event, data, err, e.batchInvokePostHandlers)

			if err != nil {
				return e.runHandleBatchInvokeError(ctx, event, err, data)
			}

			return &Result{
				Data:  data,
				Error: nil,
			}, nil
		}

		err := errors.InternalErrorf("FIELD_NOT_FOUND", "Not found handler on field %s", event.Field)
		e.runHandleBatchInvokeError(ctx, event, err, nil)

		return nil, err
	case eventInvokeType:
		event := req.InvokeEvent
		if mainHandler, ok := e.invokeFields[event.Field]; ok {
			if err := e.runInvokePreHandler(ctx, event, e.invokePreHandlers); err != nil {
				return e.runHandleInvokeError(ctx, event, err, nil)
			}

			if err := e.runInvokePreHandler(ctx, event, mainHandler.preHandlers); err != nil {
				return e.runHandleInvokeError(ctx, event, err, nil)
			}

			data, err := mainHandler.handler(ctx, event)

			e.runInvokePostHandler(ctx, event, data, err, mainHandler.postHandlers)
			e.runInvokePostHandler(ctx, event, data, err, e.invokePostHandlers)

			if err != nil {
				return e.runHandleInvokeError(ctx, event, err, data)
			}

			return &Result{
				Data:  data,
				Error: nil,
			}, nil
		}

		err := errors.InternalErrorf("FIELD_NOT_FOUND", "Not found handler on field %s", event.Field)
		e.runHandleInvokeError(ctx, event, err, nil)

		return nil, err
	}

	return nil, errors.InternalErrorf("FIELD_NOT_FOUND", "Not found handler")
}
