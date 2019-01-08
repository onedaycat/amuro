package cognitoevent

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type CognitoPreSignupEventHandler func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error)
type CognitoPreSignupPreHandler func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup)
type CognitoPreSignupPostHandler func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup, err error)

type CognitoPreSignupMainHandler struct {
	preHandlers  []CognitoPreSignupPreHandler
	postHandlers []CognitoPreSignupPostHandler
	handler      CognitoPreSignupEventHandler
}

func (e *EventManager) RegisterPreSignupHandlers(handler CognitoPreSignupEventHandler, options ...OptionPreSignup) {
	opts := newOptionPreSignup(options...)

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

func WithPreSignupPreHandlers(handlers ...CognitoPreSignupPreHandler) OptionPreSignup {
	return func(o *optionPreSignup) {
		o.preHandlers = handlers
	}
}

func WithPreSignupPostHandlers(handlers ...CognitoPreSignupPostHandler) OptionPreSignup {
	return func(o *optionPreSignup) {
		o.postHandlers = handlers
	}
}

type OptionPreSignup func(o *optionPreSignup)

type optionPreSignup struct {
	preHandlers  []CognitoPreSignupPreHandler
	postHandlers []CognitoPreSignupPostHandler
}

func newOptionPreSignup(opts ...OptionPreSignup) *optionPreSignup {
	o := &optionPreSignup{}
	if opts == nil {
		return o
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
