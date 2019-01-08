package cognitoevent

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/onedaycat/errors"
	"github.com/stretchr/testify/require"
)

func TestHandlerPostConfirmation(t *testing.T) {
	confirmationHandlerCheck := false

	mainPostHanlder := NewCognitoPostConfirmationMainHandler(func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
		confirmationHandlerCheck = true
		return event, nil
	}, nil, nil)

	eventManager := NewEventManager()
	eventManager.RegisterPostConfirmationHandlers(mainPostHanlder, nil, nil)

	requestEvent := events.CognitoEventUserPoolsPostConfirmation{}
	responseEvent, err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, confirmationHandlerCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestHandlerPreSignup(t *testing.T) {
	preSignupHandlerCheck := false

	mainPreHanlder := NewCognitoPreSignupMainHandler(func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
		preSignupHandlerCheck = true
		return event, nil
	}, nil, nil)

	eventManager := NewEventManager()
	eventManager.RegisterPreSignupHandlers(mainPreHanlder, nil, nil)

	requestEvent := events.CognitoEventUserPoolsPreSignup{}
	responseEvent, err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, preSignupHandlerCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestCustomErrorHandle(t *testing.T) {
	errorHandleCheck := false

	mainPreHanlder := NewCognitoPreSignupMainHandler(func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
		return event, errors.InternalError("test_custom_error_handle", "test")
	}, nil, nil)

	eventManager := NewEventManager()
	eventManager.RegisterPreSignupHandlers(mainPreHanlder, nil, nil)

	eventManager.OnError = func(ctx context.Context, event interface{}, err error) {
		require.Equal(t, events.CognitoEventUserPoolsPreSignup{}, event.(events.CognitoEventUserPoolsPreSignup))
		errorHandleCheck = true
	}

	requestEvent := events.CognitoEventUserPoolsPreSignup{}
	responseEvent, err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Error(t, err)
	require.True(t, errorHandleCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestFlowPostConfirmationHandlers(t *testing.T) {
	var mainHandlerCheck, mainPreHanlderCheck, mainPostHandlerCheck, eventPreHandlerCheck, eventPostHandlerCheck bool

	mainPostHanlder := NewCognitoPostConfirmationMainHandler(
		func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
			mainHandlerCheck = true
			return event, nil
		}, []CognitoPostConfirmationPreHandler{
			func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) {
				mainPreHanlderCheck = true
			},
		},
		[]CognitoPostConfirmationPostHandler{
			func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation, err error) {
				mainPostHandlerCheck = true
			},
		},
	)

	eventManager := NewEventManager()
	eventManager.RegisterPostConfirmationHandlers(
		mainPostHanlder,
		[]CognitoPostConfirmationPreHandler{
			func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) {
				eventPreHandlerCheck = true
			},
		}, []CognitoPostConfirmationPostHandler{
			func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation, err error) {
				eventPostHandlerCheck = true
			},
		},
	)

	requestEvent := events.CognitoEventUserPoolsPostConfirmation{}
	responseEvent, err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, mainHandlerCheck)
	require.True(t, mainPreHanlderCheck)
	require.True(t, mainPostHandlerCheck)
	require.True(t, eventPreHandlerCheck)
	require.True(t, eventPostHandlerCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestFlowPreSignupHandlers(t *testing.T) {
	var mainHandlerCheck, mainPreHanlderCheck, mainPostHandlerCheck, eventPreHandlerCheck, eventPostHandlerCheck bool

	mainPreHanlder := NewCognitoPreSignupMainHandler(
		func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
			mainHandlerCheck = true
			return event, nil
		}, []CognitoPreSignupPreHandler{
			func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) {
				mainPreHanlderCheck = true
			},
		},
		[]CognitoPreSignupPostHandler{
			func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup, err error) {
				mainPostHandlerCheck = true
			},
		},
	)

	eventManager := NewEventManager()
	eventManager.RegisterPreSignupHandlers(
		mainPreHanlder,
		[]CognitoPreSignupPreHandler{
			func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) {
				eventPreHandlerCheck = true
			},
		}, []CognitoPreSignupPostHandler{
			func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup, err error) {
				eventPostHandlerCheck = true
			},
		},
	)

	requestEvent := events.CognitoEventUserPoolsPreSignup{}
	responseEvent, err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, mainHandlerCheck)
	require.True(t, mainPreHanlderCheck)
	require.True(t, mainPostHandlerCheck)
	require.True(t, eventPreHandlerCheck)
	require.True(t, eventPostHandlerCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestFlowMultiTypeHandlers(t *testing.T) {
	var preSignupHandlerCheck, postConfirmationHandlerCheck bool

	preHandler := NewCognitoPreSignupMainHandler(func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
		preSignupHandlerCheck = true
		return event, nil
	}, nil, nil)

	postHanlder := NewCognitoPostConfirmationMainHandler(func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
		postConfirmationHandlerCheck = true
		return event, nil
	}, nil, nil)

	eventManager := NewEventManager()
	eventManager.RegisterPreSignupHandlers(preHandler, nil, nil)
	eventManager.RegisterPostConfirmationHandlers(postHanlder, nil, nil)

	requestEvent := events.CognitoEventUserPoolsPreSignup{}
	responseEvent, err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, preSignupHandlerCheck)
	require.False(t, postConfirmationHandlerCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestNotImplmentEvent(t *testing.T) {
	requestEvent := events.CognitoEventUserPoolsPreTokenGen{}

	eventManager := NewEventManager()
	responseEvent, err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Error(t, err)
	require.Equal(t, "HANDLER_NOT_FOUND: Not found handler on event: events.CognitoEventUserPoolsPreTokenGen", err.Error())
	require.Equal(t, requestEvent, responseEvent)
}
