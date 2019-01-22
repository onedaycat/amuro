package appsync

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/onedaycat/errors"
	"github.com/stretchr/testify/require"
)

func TestParseBatchInvokeEvent(t *testing.T) {
	testcases := []struct {
		payload  string
		expEvent *Event
	}{
		{
			`[{"field": "testField1","arguments": {"arg1": "1"},"source": { "namespace": "1" },"identity": {"sub": "xx"}},
			{"field": "testField1","arguments": {"arg1": "1"},"source": { "namespace": "2" },"identity": {"sub": "xx"}}]`,
			&Event{"testField1", []byte(`{"arg1": "1"}`), nil, &Identity{Sub: "xx"}, []map[string]interface{}{{"namespace": "1"}, {"namespace": "2"}}},
		},
		// no field
		{
			`[{"arguments": {"arg1": "1"},"source": { "namespace": "1" },"identity": {"sub": "xx"}},
			{"field": "testField1","arguments": {"arg1": "1"},"source": { "namespace": "2" },"identity": {"sub": "xx"}}]`,
			&Event{"", []byte(`{"arg1": "1"}`), nil, &Identity{Sub: "xx"}, []map[string]interface{}{{"namespace": "1"}, {"namespace": "2"}}},
		},
		// no args
		{
			`[{"field": "testField1","source": { "namespace": "1" },"identity": {"sub": "xx"}},
			{"field": "testField1","arguments": {"arg1": "1"},"source": { "namespace": "2" },"identity": {"sub": "xx"}}]`,
			&Event{"testField1", nil, nil, &Identity{Sub: "xx"}, []map[string]interface{}{{"namespace": "1"}, {"namespace": "2"}}},
		},
		// no identity
		{
			`[{"field": "testField1","arguments": {"arg1": "1"},"source": { "namespace": "1" }},
			{"field": "testField1","arguments": {"arg1": "1"},"source": { "namespace": "2" },"identity": {"sub": "xx"}}]`,
			&Event{"testField1", []byte(`{"arg1": "1"}`), nil, nil, []map[string]interface{}{{"namespace": "1"}, {"namespace": "2"}}},
		},
		// missing source 1
		{
			`[{"field": "testField1","arguments": {"arg1": "1"},"identity": {"sub": "xx"}},
			{"field": "testField1","arguments": {"arg1": "1"},"source": { "namespace": "2" },"identity": {"sub": "xx"}}]`,
			&Event{"testField1", []byte(`{"arg1": "1"}`), nil, &Identity{Sub: "xx"}, []map[string]interface{}{{"namespace": "2"}}},
		},
		// missing source 2
		{
			`[{"field": "testField1","arguments": {"arg1": "1"},"source": { "namespace": "1" },"identity": {"sub": "xx"}},
			{"field": "testField1","arguments": {"arg1": "1"},"identity": {"sub": "xx"}}]`,
			&Event{"testField1", []byte(`{"arg1": "1"}`), nil, &Identity{Sub: "xx"}, []map[string]interface{}{{"namespace": "1"}}},
		},
		// no source
		{
			`[{"field": "testField1","arguments": {"arg1": "1"},"identity": {"sub": "xx"}},
			{"field": "testField1","arguments": {"arg1": "1"},"identity": {"sub": "xx"}}]`,
			&Event{"testField1", []byte(`{"arg1": "1"}`), nil, &Identity{Sub: "xx"}, nil},
		},
	}

	for i, testcase := range testcases {
		event := &Event{}
		err := json.Unmarshal([]byte(testcase.payload), event)
		require.NoError(t, err)
		require.Equal(t, testcase.expEvent, event, i)
	}
}

func TestParseInvokeEvent(t *testing.T) {
	testcases := []struct {
		payload  string
		expEvent *Event
	}{
		{
			`{"field": "testField1","arguments": {"arg1": "1"},"source": {"namespace": "1"},"identity": {"sub": "xx"}}`,
			&Event{"testField1", []byte(`{"arg1": "1"}`), []byte(`{"namespace": "1"}`), &Identity{Sub: "xx"}, nil},
		},
		// no field
		{
			`{"arguments": {"arg1": "1"},"source": {"namespace": "1"},"identity": {"sub": "xx"}}`,
			&Event{"", []byte(`{"arg1": "1"}`), []byte(`{"namespace": "1"}`), &Identity{Sub: "xx"}, nil},
		},
		// no args
		{
			`{"field": "testField1","source": {"namespace": "1"},"identity": {"sub": "xx"}}`,
			&Event{"testField1", nil, []byte(`{"namespace": "1"}`), &Identity{Sub: "xx"}, nil},
		},
		// no identity
		{
			`{"field": "testField1","arguments": {"arg1": "1"},"source": {"namespace": "1"}}`,
			&Event{"testField1", []byte(`{"arg1": "1"}`), []byte(`{"namespace": "1"}`), nil, nil},
		},
		// no source
		{
			`{"field": "testField1","arguments": {"arg1": "1"},"identity": {"sub": "xx"}}`,
			&Event{"testField1", []byte(`{"arg1": "1"}`), nil, &Identity{Sub: "xx"}, nil},
		},
	}

	for i, testcase := range testcases {
		event := &Event{}
		err := json.Unmarshal([]byte(testcase.payload), event)
		require.NoError(t, err)
		require.Equal(t, testcase.expEvent, event, i)
	}
}

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
	require.Nil(t, err)
	require.Equal(t, 3, triggerCount)
	require.Equal(t, string(expectedResultRawJSON), string(result.Data.(json.RawMessage)))
	require.Equal(t, string(postFuncResult), string(result.Data.(json.RawMessage)))

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
	resultRawJSON := result.Data.(json.RawMessage)

	require.Nil(t, err)
	require.Equal(t, 3, triggerCount)
	require.Equal(t, string(expectedResultRawJSON), string(resultRawJSON))
	require.Equal(t, string(preInputData), string(resultRawJSON))
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
	require.Nil(t, err)
	require.Equal(t, 2, triggerCount)
	require.Equal(t, "TEST_ERROR: ERROR_AT_PRE_MAIN_HANDLE", result.Error.Error())
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
	require.Equal(t, string(expectedResultRawJSON), string(resultRawJSON))
	require.Equal(t, string(expectedResultRawJSON), string(postInput))
	require.Equal(t, string(expectedResultRawJSON), string(postResult.(json.RawMessage)))
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

	require.Nil(t, err)
	require.Equal(t, expectedResult, result)
	require.Equal(t, 1, triggerCount)
}
