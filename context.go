package appsync

import (
	"context"
)

type customContext struct {
	handlers  []EventHandler
	eventData *Event
	context.Context
	result              interface{}
	handlerlsLength     int
	currentHandlerIndex int
}

type EventContext interface {
	context.Context

	Next() (interface{}, error)
	IsStopped() bool
}

func NewEventContext(handlers []EventHandler, eventData *Event) EventContext {
	ct := &customContext{
		handlers:            handlers,
		eventData:           eventData,
		currentHandlerIndex: 0,
		handlerlsLength:     len(handlers),
	}

	return ct
}

func (ct *customContext) IsStopped() bool {
	ct.currentHandlerIndex++
	if ct.currentHandlerIndex < ct.handlerlsLength {
		return false
	}

	return true
}

func (ct *customContext) Next() (interface{}, error) {
	result, err := ct.handlers[ct.currentHandlerIndex](ct, ct.eventData)
	if err != nil {
		return nil, err
	}

	if result != nil {
		ct.result = result
	}

	if !ct.IsStopped() {
		ct.Next()
	}

	return ct.result, nil
}
