package cognitoevent

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
)

type ErrorHandler func(ctx context.Context, event interface{}, err error)

type EventManager struct {
	postConfirmationMainHandler *CognitoPostConfirmationMainHandler
	preSignupMainHandler        *CognitoPreSignupMainHandler

	OnError ErrorHandler
}

func NewEventManager() *EventManager {
	return &EventManager{}
}

func (e *EventManager) RegisterPreSignupHandlers(handler CognitoPreSignupEventHandler, options ...PreSignupOption) {
	opts := newPreSignupOption(options...)

	e.preSignupMainHandler = &CognitoPreSignupMainHandler{
		handler: handler,
	}

	if len(opts.preHandlers) > 0 {
		e.preSignupMainHandler.preHandlers = opts.preHandlers
	}

	if len(opts.postHandlers) > 0 {
		e.preSignupMainHandler.postHandlers = opts.postHandlers
	}
}

func (e *EventManager) RegisterPostConfirmationHandlers(handler CognitoPostConfirmationEventHandler, options ...PostConfirmationOption) {
	opts := newPostConfirmationOption(options...)

	e.postConfirmationMainHandler = &CognitoPostConfirmationMainHandler{
		handler: handler,
	}

	if len(opts.preHandlers) > 0 {
		e.postConfirmationMainHandler.preHandlers = opts.preHandlers
	}

	if len(opts.postHandlers) > 0 {
		e.postConfirmationMainHandler.postHandlers = opts.postHandlers
	}
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
	e.runPostConfirmationPreHandler(ctx, event, e.postConfirmationMainHandler.preHandlers)

	respEvent, err := e.postConfirmationMainHandler.handler(ctx, event)

	e.runPostConfirmationPostHandler(ctx, event, err, e.postConfirmationMainHandler.postHandlers)

	return respEvent, err
}

func (e *EventManager) runPreSingup(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
	e.runPreSignupPreHandler(ctx, event, e.preSignupMainHandler.preHandlers)

	respEvent, err := e.preSignupMainHandler.handler(ctx, event)

	e.runPreSignupPostHandler(ctx, event, err, e.preSignupMainHandler.postHandlers)

	return respEvent, err
}

func (e *EventManager) RunPreSignup(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
	if e.preSignupMainHandler == nil {
		return event, notImplementHandlerOnEvent("preSignup")
	}

	respEvent, err := e.runPreSingup(ctx, event)
	if err != nil && e.OnError != nil {
		e.OnError(ctx, event, err)
	}

	return respEvent, err
}

func (e *EventManager) RunPostConfirmation(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	if e.postConfirmationMainHandler == nil {
		return event, notImplementHandlerOnEvent("postConfirmation")
	}

	respEvent, err := e.runPostConfirmation(ctx, event)
	if err != nil && e.OnError != nil {
		e.OnError(ctx, event, err)
	}

	return respEvent, err
}

func notImplementHandlerOnEvent(eventType string) error {
	return errors.InternalErrorf("HANDLER_NOT_FOUND", "Not found handler on event: %s", eventType)
}
