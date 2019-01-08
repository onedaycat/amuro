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

func (e *EventManager) RegisterPreSignupHandlers(handler CognitoPreSignupEventHandler, preHandlers []CognitoPreSignupPreHandler, postHandlers []CognitoPreSignupPostHandler) {
	e.preSignupMainHandler = &CognitoPreSignupMainHandler{
		handler:      handler,
		preHandlers:  preHandlers,
		postHandlers: postHandlers,
	}
}
