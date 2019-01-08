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

func WithPreSignupPreHandlers(handlers ...CognitoPreSignupPreHandler) PreSignupOption {
	return func(o *preSignupOption) {
		o.preHandlers = handlers
	}
}

func WithPreSignupPostHandlers(handlers ...CognitoPreSignupPostHandler) PreSignupOption {
	return func(o *preSignupOption) {
		o.postHandlers = handlers
	}
}

type PreSignupOption func(o *preSignupOption)

type preSignupOption struct {
	preHandlers  []CognitoPreSignupPreHandler
	postHandlers []CognitoPreSignupPostHandler
}

func newPreSignupOption(opts ...PreSignupOption) *preSignupOption {
	o := &preSignupOption{}
	if opts == nil {
		return o
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
