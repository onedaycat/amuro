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
	eventManager.RegisterPostConfirmationHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmationRequest) error {
		confirmationHandlerCheck = true
		return nil
	})

	requestEvent := events.CognitoEventUserPoolsPostConfirmationRequest{}
	err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, confirmationHandlerCheck)
}

func TestHandlerPreSignup(t *testing.T) {
	preSignupHandlerCheck := false

	eventManager := NewEventManager()
	eventManager.RegisterPreSignupHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPreSignupRequest) error {
		preSignupHandlerCheck = true
		return nil
	})

	requestEvent := events.CognitoEventUserPoolsPreSignupRequest{}
	err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, preSignupHandlerCheck)
}

func TestCustomErrorHandle(t *testing.T) {
	errorHandleCheck := false

	eventManager := NewEventManager()
	eventManager.RegisterPreSignupHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPreSignupRequest) error {
		return errors.InternalError("test_custom_error_handle", "test")
	})

	eventManager.OnError = func(ctx context.Context, event interface{}, err error) {
		require.Equal(t, events.CognitoEventUserPoolsPreSignupRequest{}, event.(events.CognitoEventUserPoolsPreSignupRequest))
		errorHandleCheck = true
	}

	requestEvent := events.CognitoEventUserPoolsPreSignupRequest{}

	err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Error(t, err)
	require.True(t, errorHandleCheck)
}

func TestFlowPostConfirmationHandlers(t *testing.T) {
	var mainHandlerCheck, mainPreHanlderCheck, mainPostHandlerCheck, globalPreHandlerCheck, globalPostHandlerCheck bool

	eventManager := NewEventManager()

	eventManager.UsePreHandler(func(ctx context.Context, event interface{}) { globalPreHandlerCheck = true })
	eventManager.UsePostHandler(func(ctx context.Context, event interface{}, err error) { globalPostHandlerCheck = true })

	eventManager.RegisterPostConfirmationHandlers(
		func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmationRequest) error {
			mainHandlerCheck = true
			return nil
		},
		WithPostConfirmationPreHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmationRequest) {
			mainPreHanlderCheck = true
		}),
		WithPostConfirmationPostHandlers(
			func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmationRequest, err error) {
				mainPostHandlerCheck = true
			},
		),
	)

	requestEvent := events.CognitoEventUserPoolsPostConfirmationRequest{}
	err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, mainHandlerCheck)
	require.True(t, mainPreHanlderCheck)
	require.True(t, mainPostHandlerCheck)
	require.True(t, globalPreHandlerCheck)
	require.True(t, globalPostHandlerCheck)
}

func TestFlowPreSignupHandlers(t *testing.T) {
	var mainHandlerCheck, mainPreHanlderCheck, mainPostHandlerCheck, globalPreHandlerCheck, globalPostHandlerCheck bool

	eventManager := NewEventManager()
	eventManager.UsePreHandler(func(ctx context.Context, event interface{}) { globalPreHandlerCheck = true })
	eventManager.UsePostHandler(func(ctx context.Context, event interface{}, err error) { globalPostHandlerCheck = true })

	eventManager.RegisterPreSignupHandlers(
		func(ctx context.Context, event events.CognitoEventUserPoolsPreSignupRequest) error {
			mainHandlerCheck = true
			return nil
		},
		WithPreSignupPreHandlers(
			func(ctx context.Context, event events.CognitoEventUserPoolsPreSignupRequest) {
				mainPreHanlderCheck = true
			},
		),
		WithPreSignupPostHandlers(
			func(ctx context.Context, event events.CognitoEventUserPoolsPreSignupRequest, err error) {
				mainPostHandlerCheck = true
			},
		),
	)

	requestEvent := events.CognitoEventUserPoolsPreSignupRequest{}
	err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, mainHandlerCheck)
	require.True(t, mainPreHanlderCheck)
	require.True(t, mainPostHandlerCheck)
	require.True(t, globalPreHandlerCheck)
	require.True(t, globalPostHandlerCheck)
}

func TestFlowMultiTypeHandlers(t *testing.T) {
	var preSignupHandlerCheck, postConfirmationHandlerCheck bool

	eventManager := NewEventManager()
	eventManager.RegisterPreSignupHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPreSignupRequest) error {
		preSignupHandlerCheck = true
		return nil
	})
	eventManager.RegisterPostConfirmationHandlers(func(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmationRequest) error {
		postConfirmationHandlerCheck = true
		return nil
	})

	requestEvent := events.CognitoEventUserPoolsPreSignupRequest{}
	err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Nil(t, err)
	require.True(t, preSignupHandlerCheck)
	require.False(t, postConfirmationHandlerCheck)
}

func TestNotImplmentEvent(t *testing.T) {
	requestEvent := events.CognitoEventUserPoolsPreTokenGen{}
	requestEvent2 := events.CognitoEventUserPoolsPostConfirmationRequest{}

	eventManager := NewEventManager()
	err := eventManager.MainHandler(context.Background(), requestEvent)
	require.Error(t, err)
	require.Equal(t, "HANDLER_NOT_FOUND: Not found handler on event: events.CognitoEventUserPoolsPreTokenGen", err.Error())

	err = eventManager.MainHandler(context.Background(), requestEvent2)
	require.Error(t, err)
	require.Equal(t, "HANDLER_NOT_FOUND: Not found handler on event: events.CognitoEventUserPoolsPostConfirmationRequest", err.Error())
}
