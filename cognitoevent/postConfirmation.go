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

type PostConfirmationOption func(o *postConfirmationOption)

type postConfirmationOption struct {
	preHandlers  []CognitoPostConfirmationPreHandler
	postHandlers []CognitoPostConfirmationPostHandler
}

func WithPostConfirmationPreHandlers(handlers ...CognitoPostConfirmationPreHandler) PostConfirmationOption {
	return func(o *postConfirmationOption) {
		o.preHandlers = handlers
	}
}

func WithPostConfirmationPostHandlers(handlers ...CognitoPostConfirmationPostHandler) PostConfirmationOption {
	return func(o *postConfirmationOption) {
		o.postHandlers = handlers
	}
}

func newPostConfirmationOption(opts ...PostConfirmationOption) *postConfirmationOption {
	o := &postConfirmationOption{}
	if opts == nil {
		return o
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
