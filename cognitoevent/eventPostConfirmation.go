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

func NewCognitoPostConfirmationMainHandler(handler CognitoPostConfirmationEventHandler, preHandlers []CognitoPostConfirmationPreHandler, postHandlers []CognitoPostConfirmationPostHandler) *CognitoPostConfirmationMainHandler {
	return &CognitoPostConfirmationMainHandler{
		handler:      handler,
		preHandlers:  preHandlers,
		postHandlers: postHandlers,
	}
}
