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

	eventManager := NewEventManager()
	eventManager.RegisterPostConfirmationHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
		confirmationHandlerCheck = true
		return event, nil
	})

	requestEvent := events.CognitoEventUserPoolsPostConfirmation{}
	responseEvent, err := eventManager.RunPostConfirmation(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, confirmationHandlerCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestHandlerPreSignup(t *testing.T) {
	preSignupHandlerCheck := false

	eventManager := NewEventManager()
	eventManager.RegisterPreSignupHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
		preSignupHandlerCheck = true
		return event, nil
	})

	requestEvent := events.CognitoEventUserPoolsPreSignup{}
	responseEvent, err := eventManager.RunPreSignup(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, preSignupHandlerCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestCustomErrorHandle(t *testing.T) {
	errorHandleCheck := false

	eventManager := NewEventManager()
	eventManager.RegisterPreSignupHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
		return event, errors.InternalError("test_custom_error_handle", "test")
	})

	eventManager.OnError = func(ctx context.Context, event interface{}, err error) {
		require.Equal(t, events.CognitoEventUserPoolsPreSignup{}, event.(events.CognitoEventUserPoolsPreSignup))
		errorHandleCheck = true
	}

	requestEvent := events.CognitoEventUserPoolsPreSignup{}

	responseEvent, err := eventManager.RunPreSignup(context.Background(), requestEvent)
	require.Error(t, err)
	require.True(t, errorHandleCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestFlowPostConfirmationHandlers(t *testing.T) {
	var mainHandlerCheck, mainPreHanlderCheck, mainPostHandlerCheck bool

	eventManager := NewEventManager()

	eventManager.RegisterPostConfirmationHandlers(
		func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
			mainHandlerCheck = true
			return event, nil
		},
		WithPostConfirmationPreHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) {
			mainPreHanlderCheck = true
		}),
		WithPostConfirmationPostHandlers(
			func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation, err error) {
				mainPostHandlerCheck = true
			},
		),
	)

	requestEvent := events.CognitoEventUserPoolsPostConfirmation{}
	responseEvent, err := eventManager.RunPostConfirmation(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, mainHandlerCheck)
	require.True(t, mainPreHanlderCheck)
	require.True(t, mainPostHandlerCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestFlowPreSignupHandlers(t *testing.T) {
	var mainHandlerCheck, mainPreHanlderCheck, mainPostHandlerCheck bool

	eventManager := NewEventManager()

	eventManager.RegisterPreSignupHandlers(
		func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
			mainHandlerCheck = true
			return event, nil
		},
		WithPreSignupPreHandlers(
			func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) {
				mainPreHanlderCheck = true
			},
		),
		WithPreSignupPostHandlers(
			func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup, err error) {
				mainPostHandlerCheck = true
			},
		),
	)

	requestEvent := events.CognitoEventUserPoolsPreSignup{}
	responseEvent, err := eventManager.RunPreSignup(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, mainHandlerCheck)
	require.True(t, mainPreHanlderCheck)
	require.True(t, mainPostHandlerCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestFlowMultiTypeHandlers(t *testing.T) {
	var preSignupHandlerCheck, postConfirmationHandlerCheck bool

	eventManager := NewEventManager()
	eventManager.RegisterPreSignupHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
		preSignupHandlerCheck = true
		return event, nil
	})
	eventManager.RegisterPostConfirmationHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
		postConfirmationHandlerCheck = true
		return event, nil
	})

	requestEvent := events.CognitoEventUserPoolsPreSignup{}
	responseEvent, err := eventManager.RunPreSignup(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, preSignupHandlerCheck)
	require.False(t, postConfirmationHandlerCheck)
	require.Equal(t, requestEvent, responseEvent)
}

func TestNotImplmentEvent(t *testing.T) {
	requestEvent2 := events.CognitoEventUserPoolsPreSignup{}

	eventManager := NewEventManager()
	responseEvent, err := eventManager.RunPreSignup(context.Background(), requestEvent2)
	require.Error(t, err)
	require.Equal(t, "HANDLER_NOT_FOUND: Not found handler on event: preSignup", err.Error())
	require.Equal(t, requestEvent2, responseEvent)
}
