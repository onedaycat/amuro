package appsync

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/onedaycat/errors"
)

func TestMainHandlerAndPassDataToMainPostHandler(t *testing.T) {
	triggerCount := 0
	expectedResultRawJSON := json.RawMessage(`{id:"id_test", check: 88}`)
	postFuncResult := []byte{}
	inputData := &Event{
		Field: "manyfunctions",
	}

	testMainHandlers := &MainHandler{
		preHandlers: []PreHandler{
			func(ctx context.Context, event *Event) error {
				triggerCount++
				return nil
			},
		},
		handler: func(ctx context.Context, event *Event) (interface{}, error) {
			triggerCount++
			return json.RawMessage(`{id:"id_test", check: 88}`), nil
		},
		postHandlers: []PostHandler{
			func(ctx context.Context, event *Event, result interface{}, err error) {
				triggerCount++
				postFuncResult = result.(json.RawMessage)
			},
		},
	}

	eventManager := NewEventManager()
	eventManager.RegisterField("manyfunctions", testMainHandlers.handler, testMainHandlers.preHandlers, testMainHandlers.postHandlers)

	result, err := eventManager.Run(context.Background(), inputData)
	if err != nil {
		t.Errorf("Next return err %s", err.Error())
	}

	if triggerCount != 3 {
		t.Errorf("Handlers not trigger")
	}

	resultRawJSON := result.Data.(json.RawMessage)
	if string(expectedResultRawJSON) != string(resultRawJSON) {
		t.Errorf("RawJSON not equal expected result")
	}

	if string(postFuncResult) != string(resultRawJSON) {
		t.Errorf("RawPostJSON not equal expected result")
	}
}

func TestPreMainHandlerTransformInput(t *testing.T) {
	triggerCount := 0
	inputData := &Event{
		Field: "testprefunctions",
		Args:  json.RawMessage(`{input: "no_edit"}`),
	}
	expectedResultRawJSON := json.RawMessage(`{input: "edited"}`)
	preInputData := []byte{}

	testMainHandlers := &MainHandler{
		preHandlers: []PreHandler{
			func(ctx context.Context, event *Event) error {
				triggerCount++
				event.Args = json.RawMessage(`{input: "edited"}`)
				return nil
			},
		},
		handler: func(ctx context.Context, event *Event) (interface{}, error) {
			triggerCount++
			return event.Args, nil
		},
		postHandlers: []PostHandler{
			func(ctx context.Context, event *Event, result interface{}, err error) {
				triggerCount++
				preInputData = event.Args
			},
		},
	}

	eventManager := NewEventManager()
	eventManager.RegisterField("testprefunctions", testMainHandlers.handler, testMainHandlers.preHandlers, testMainHandlers.postHandlers)

	result, err := eventManager.Run(context.Background(), inputData)
	if err != nil {
		t.Errorf("Next return err %s", err.Error())
	}

	if triggerCount != 3 {
		t.Errorf("Handlers not trigger")
	}

	resultRawJSON := result.Data.(json.RawMessage)
	if string(expectedResultRawJSON) != string(resultRawJSON) {
		t.Errorf("RawJSON not equal expected result")
	}

	if string(preInputData) != string(resultRawJSON) {
		t.Errorf("RawPostJSON not equal expected result")
	}
}

func TestPreMainHandlerError(t *testing.T) {
	triggerCount := 0
	inputData := &Event{
		Field: "testPreError",
		Args:  json.RawMessage(`{input: "no_edit"}`),
	}
	testMainHandlers := &MainHandler{
		preHandlers: []PreHandler{
			func(ctx context.Context, event *Event) error {
				triggerCount++
				return nil
			},
			func(ctx context.Context, event *Event) error {
				triggerCount++
				return errors.InternalErrorf("TEST_ERROR", "ERROR_AT_PRE_MAIN_HANDLE")
			},
		},
		handler: func(ctx context.Context, event *Event) (interface{}, error) {
			triggerCount++
			return event.Args, nil
		},
		postHandlers: []PostHandler{
			func(ctx context.Context, event *Event, result interface{}, err error) {
				triggerCount++
			},
		},
	}

	eventManager := NewEventManager()
	eventManager.RegisterField("testPreError", testMainHandlers.handler, testMainHandlers.preHandlers, testMainHandlers.postHandlers)

	result, err := eventManager.Run(context.Background(), inputData)
	if err != nil {
		t.Errorf("Error be nil")
	}

	if triggerCount != 2 {
		t.Errorf("Handlers should be run only preHandlerห")
	}

	if result.Error.Error() != "TEST_ERROR: ERROR_AT_PRE_MAIN_HANDLE" {
		t.Errorf("Raw result error not equal expected")
	}
}

func TestTransformInputDataAtPreHandlersAndPassToPostHandlers(t *testing.T) {
	triggerCount := 0
	inputData := &Event{
		Field: "transFromPreInputToPostHandlers",
		Args:  json.RawMessage(`{input: "no_edit"}`),
	}
	expectedResultRawJSON := json.RawMessage(`{input: "edited"}`)
	postInput := []byte{}
	var postResult interface{}

	testPreHandlers := []PreHandler{
		func(ctx context.Context, event *Event) error {
			triggerCount++
			event.Args = json.RawMessage(`{input: "edited"}`)
			return nil
		},
		func(ctx context.Context, event *Event) error {
			triggerCount++
			return nil
		},
	}
	testMainHandlers := &MainHandler{
		preHandlers: []PreHandler{
			func(ctx context.Context, event *Event) error {
				triggerCount++
				return nil
			},
			func(ctx context.Context, event *Event) error {
				triggerCount++
				return nil
			},
		},
		handler: func(ctx context.Context, event *Event) (interface{}, error) {
			triggerCount++
			return event.Args, nil
		},
		postHandlers: []PostHandler{
			func(ctx context.Context, event *Event, result interface{}, err error) {
				triggerCount++
			},
		},
	}
	testPostHandlers := []PostHandler{
		func(ctx context.Context, event *Event, result interface{}, err error) {
			triggerCount++

		},
		func(ctx context.Context, event *Event, result interface{}, err error) {
			triggerCount++
			postInput = event.Args
			postResult = result
		},
	}

	eventManager := NewEventManager()
	eventManager.UsePreHandler(testPreHandlers...)
	eventManager.UsePostHandler(testPostHandlers...)
	eventManager.RegisterField("transFromPreInputToPostHandlers", testMainHandlers.handler, testMainHandlers.preHandlers, testMainHandlers.postHandlers)

	result, err := eventManager.Run(context.Background(), inputData)
	if err != nil {
		t.Errorf("Error be nil")
	}

	if triggerCount != 8 {
		t.Errorf("Should be run all handlers")
	}

	resultRawJSON := result.Data.(json.RawMessage)
	if string(expectedResultRawJSON) != string(resultRawJSON) {
		t.Errorf("RawJSON not equal expected result")
	}

	if string(postInput) != string(expectedResultRawJSON) {
		t.Errorf("PostInput not equal expected result")
	}

	if string(postResult.(json.RawMessage)) != string(expectedResultRawJSON) {
		t.Errorf("PostInput not equal expected result")
	}
}

func TestRunOnError(t *testing.T) {
	triggerCount := 0
	inputData := &Event{
		Field: "testError",
	}

	expectedResult := &Result{
		Data:  nil,
		Error: errors.InternalError("UNKNOWN_CODE", "Test unknown error"),
	}

	testMainHandlers := &MainHandler{
		preHandlers: []PreHandler{},
		handler: func(ctx context.Context, event *Event) (interface{}, error) {
			return nil, errors.InternalError("UNKNOWN_CODE", "Test unknown error")
		},
		postHandlers: []PostHandler{},
	}

	errorFunc := func(ctx context.Context, event *Event, err error) {
		triggerCount++
	}

	eventManager := NewEventManager()
	eventManager.OnError(errorFunc)
	eventManager.RegisterField("testError", testMainHandlers.handler, testMainHandlers.preHandlers, testMainHandlers.postHandlers)

	result, err := eventManager.Run(context.Background(), inputData)
	if err == nil {
		t.Errorf("Error should not be nil")
	}

	if result == expectedResult {
		t.Errorf("Result should be equal expected")
	}

	if triggerCount != 1 {
		t.Errorf("Should run only trigger error")
	}
}
