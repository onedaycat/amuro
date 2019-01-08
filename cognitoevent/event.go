package cognitoevent

import (
	"context"
	"reflect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
)

type ErrorHandler func(ctx context.Context, event interface{}, err error)

type EventManager struct {
	postConfirmationPreHandlers  []CognitoPostConfirmationPreHandler
	postConfirmationPostHandlers []CognitoPostConfirmationPostHandler
	postConfirmationMainHandler  *CognitoPostConfirmationMainHandler

	preSignupPreHandlers  []CognitoPreSignupPreHandler
	preSignupPostHandlers []CognitoPreSignupPostHandler
	preSignupMainHandler  *CognitoPreSignupMainHandler

	OnError ErrorHandler
}

func NewEventManager() *EventManager {
	return &EventManager{
		postConfirmationPreHandlers:  []CognitoPostConfirmationPreHandler{},
		postConfirmationPostHandlers: []CognitoPostConfirmationPostHandler{},
		postConfirmationMainHandler:  &CognitoPostConfirmationMainHandler{},
		preSignupPreHandlers:         []CognitoPreSignupPreHandler{},
		preSignupPostHandlers:        []CognitoPreSignupPostHandler{},
		preSignupMainHandler:         &CognitoPreSignupMainHandler{},
	}
}

func (e *EventManager) RegisterPostConfirmationHandlers(mainHandler *CognitoPostConfirmationMainHandler, preHandlers []CognitoPostConfirmationPreHandler, postHandlers []CognitoPostConfirmationPostHandler) {
	e.postConfirmationMainHandler = mainHandler
	e.postConfirmationPreHandlers = preHandlers
	e.postConfirmationPostHandlers = postHandlers
}

func (e *EventManager) RegisterPreSignupHandlers(mainHandler *CognitoPreSignupMainHandler, preHandlers []CognitoPreSignupPreHandler, postHandlers []CognitoPreSignupPostHandler) {
	e.preSignupMainHandler = mainHandler
	e.preSignupPreHandlers = preHandlers
	e.preSignupPostHandlers = postHandlers
}

func (e *EventManager) runPostConfirmationPreHandler(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation, handlers []CognitoPostConfirmationPreHandler) {
	for _, handler := range handlers {
		handler(ctx, event)
	}
}

func (e *EventManager) runPostConfirmationPostHandler(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation, handlerErr error, handlers []CognitoPostConfirmationPostHandler) {
	for _, handler := range handlers {
		handler(ctx, event, handlerErr)
	}
}

func (e *EventManager) runPreSignupPreHandler(ctx context.Context, event events.CognitoEventUserPoolsPreSignup, handlers []CognitoPreSignupPreHandler) {
	for _, handler := range handlers {
		handler(ctx, event)
	}
}

func (e *EventManager) runPreSignupPostHandler(ctx context.Context, event events.CognitoEventUserPoolsPreSignup, handlerErr error, handlers []CognitoPreSignupPostHandler) {
	for _, handler := range handlers {
		handler(ctx, event, handlerErr)
	}
}

func (e *EventManager) runPostConfirmation(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	e.runPostConfirmationPreHandler(ctx, event, e.postConfirmationPreHandlers)
	e.runPostConfirmationPreHandler(ctx, event, e.postConfirmationMainHandler.preHandlers)

	respEvent, err := e.postConfirmationMainHandler.handler(ctx, event)

	e.runPostConfirmationPostHandler(ctx, event, err, e.postConfirmationMainHandler.postHandlers)
	e.runPostConfirmationPostHandler(ctx, event, err, e.postConfirmationPostHandlers)

	return respEvent, err
}

func (e *EventManager) runPreSingup(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
	e.runPreSignupPreHandler(ctx, event, e.preSignupPreHandlers)
	e.runPreSignupPreHandler(ctx, event, e.preSignupMainHandler.preHandlers)

	respEvent, err := e.preSignupMainHandler.handler(ctx, event)

	e.runPreSignupPostHandler(ctx, event, err, e.preSignupMainHandler.postHandlers)
	e.runPreSignupPostHandler(ctx, event, err, e.preSignupPostHandlers)

	return respEvent, err
}

func (e *EventManager) MainHandler(ctx context.Context, event interface{}) (interface{}, error) {
	respEvent, err := e.Run(ctx, event)
	if err != nil && e.OnError != nil {
		e.OnError(ctx, event, err)
	}

	return respEvent, err
}

func (e *EventManager) Run(ctx context.Context, event interface{}) (interface{}, error) {
	switch v := event.(type) {
	case events.CognitoEventUserPoolsPostConfirmation:
		return e.runPostConfirmation(ctx, v)
	case events.CognitoEventUserPoolsPreSignup:
		return e.runPreSingup(ctx, v)
	default:
		return event, errors.InternalErrorf("HANDLER_NOT_FOUND", "Not found handler on event: %v", reflect.TypeOf(event))
	}
}
