package cognitoevent

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type CognitoPostConfirmationEventHandler func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error)
type CognitoPostConfirmationPreHandler func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation)
type CognitoPostConfirmationPostHandler func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation, err error)

type CognitoPostConfirmationMainHandler struct {
	preHandlers  []CognitoPostConfirmationPreHandler
	postHandlers []CognitoPostConfirmationPostHandler
	handler      CognitoPostConfirmationEventHandler
}

func (e *EventManager) RegisterPostConfirmationHandlers(handler CognitoPostConfirmationEventHandler, options ...OptionPostConfirmation) {
	opts := newOptionPostConfirmation(options...)

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

type OptionPostConfirmation func(o *optionPostConfirmation)

type optionPostConfirmation struct {
	preHandlers  []CognitoPostConfirmationPreHandler
	postHandlers []CognitoPostConfirmationPostHandler
}

func WithPostConfirmationPreHandlers(handlers ...CognitoPostConfirmationPreHandler) OptionPostConfirmation {
	return func(o *optionPostConfirmation) {
		o.preHandlers = handlers
	}
}

func WithPostConfirmationPostHandlers(handlers ...CognitoPostConfirmationPostHandler) OptionPostConfirmation {
	return func(o *optionPostConfirmation) {
		o.postHandlers = handlers
	}
}

func newOptionPostConfirmation(opts ...OptionPostConfirmation) *optionPostConfirmation {
	o := &optionPostConfirmation{}
	if opts == nil {
		return o
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
