package cognitoevent

import (
	"context"
	"reflect"

	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
)

type ErrorHandler func(ctx context.Context, event interface{}, err error)
type PreHandler func(ctx context.Context, event interface{})
type PostHandler func(ctx context.Context, event interface{}, err error)

type EventManager struct {
	preHandlers  []PreHandler
	postHandlers []PostHandler

	postConfirmationMainHandler *CognitoPostConfirmationMainHandler
	preSignupMainHandler        *CognitoPreSignupMainHandler

	OnError ErrorHandler
}

func NewEventManager(options ...Option) *EventManager {
	opts := newOption(options...)
	eventManager := &EventManager{
		preHandlers:  []PreHandler{},
		postHandlers: []PostHandler{},
	}

	if len(opts.preHandlers) > 0 {
		eventManager.preHandlers = opts.preHandlers
	}

	if len(opts.postHandlers) > 0 {
		eventManager.postHandlers = opts.postHandlers
	}

	return eventManager
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

func (e *EventManager) runGlobalPreHandler(ctx context.Context, event interface{}, handlers []PreHandler) {
	for _, handler := range handlers {
		handler(ctx, event)
	}
}

func (e *EventManager) runGlobalPostHandler(ctx context.Context, event interface{}, handlerErr error, handlers []PostHandler) {
	for _, handler := range handlers {
		handler(ctx, event, handlerErr)
	}
}

func (e *EventManager) runPostConfirmation(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	e.runGlobalPreHandler(ctx, event, e.preHandlers)
	e.runPostConfirmationPreHandler(ctx, event, e.postConfirmationMainHandler.preHandlers)

	respEvent, err := e.postConfirmationMainHandler.handler(ctx, event)

	e.runPostConfirmationPostHandler(ctx, event, err, e.postConfirmationMainHandler.postHandlers)
	e.runGlobalPostHandler(ctx, event, err, e.postHandlers)

	return respEvent, err
}

func (e *EventManager) runPreSingup(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
	e.runGlobalPreHandler(ctx, event, e.preHandlers)
	e.runPreSignupPreHandler(ctx, event, e.preSignupMainHandler.preHandlers)

	respEvent, err := e.preSignupMainHandler.handler(ctx, event)

	e.runPreSignupPostHandler(ctx, event, err, e.preSignupMainHandler.postHandlers)
	e.runGlobalPostHandler(ctx, event, err, e.postHandlers)

	return respEvent, err
}

func (e *EventManager) MainHandler(ctx context.Context, event interface{}) (interface{}, error) {
	respEvent, err := e.run(ctx, event)
	if err != nil && e.OnError != nil {
		e.OnError(ctx, event, err)
	}

	return respEvent, err
}

func (e *EventManager) run(ctx context.Context, event interface{}) (interface{}, error) {
	switch v := event.(type) {
	case events.CognitoEventUserPoolsPostConfirmation:
		if e.postConfirmationMainHandler == nil {
			return notImplementHandlerOnEvent(event)
		}

		return e.runPostConfirmation(ctx, v)
	case events.CognitoEventUserPoolsPreSignup:
		if e.preSignupMainHandler == nil {
			return notImplementHandlerOnEvent(event)
		}

		return e.runPreSingup(ctx, v)
	default:
		return notImplementHandlerOnEvent(event)
	}
}

func notImplementHandlerOnEvent(event interface{}) (interface{}, error) {
	return event, errors.InternalErrorf("HANDLER_NOT_FOUND", "Not found handler on event: %v", reflect.TypeOf(event))
}
