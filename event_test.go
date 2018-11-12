package appsync

import (
	"context"
	"encoding/json"
	"testing"
)

var triggerCount = 0

var testMainHandlers = &MainHandler{
	preHandlers: []PreHandler{
		func(ctx context.Context, event *Event) error {
			triggerCount++
			return nil
		},
	},
	postHandlers: []PostHandler{},
	handler: func(ctx context.Context, event *Event) (interface{}, error) {
		triggerCount++
		return json.RawMessage(`{id:"id_test", check: 100}`), nil
	},
}

func TestMainHandlerAndRun(t *testing.T) {
	triggerCount = 0
	expectedRawJSON := json.RawMessage(`{id:"id_test", check: 100}`)

	eventData := &Event{
		Field: "manyfunctions",
	}

	eventManager := NewEventManager()
	eventManager.RegisterField("manyfunctions", testMainHandlers.handler, testMainHandlers.preHandlers, testMainHandlers.postHandlers)

	result, err := eventManager.Run(context.Background(), eventData)
	if err != nil {
		t.Errorf("Next return err %s", err.Error())
	}

	if triggerCount != 2 {
		t.Errorf("Handlers not trigger")
	}

	resultRawJSON := result.Data.(json.RawMessage)
	if string(expectedRawJSON) != string(resultRawJSON) {
		t.Errorf("RawJSON not equal expected result")
	}
}

func TestPreHandler(t *testing.T) {
	triggerCount = 0
	expectedRawJSON := json.RawMessage(`{id:"id_test", check: 100}`)

	eventData := &Event{
		Field: "manyfunctions",
	}

	eventManager := NewEventManager()
	eventManager.UsePreHandler(
		func(ctx context.Context, event *Event) error {
			triggerCount++
			return nil
		},
		func(ctx context.Context, event *Event) error {
			triggerCount++
			return nil
		},
	)
	eventManager.RegisterField("manyfunctions", testMainHandlers.handler, testMainHandlers.preHandlers, testMainHandlers.postHandlers)

	result, err := eventManager.Run(context.Background(), eventData)
	if err != nil {
		t.Errorf("Next return err %s", err.Error())
	}

	if triggerCount != 4 {
		t.Errorf("Handlers not trigger")
	}

	resultRawJSON := result.Data.(json.RawMessage)
	if string(expectedRawJSON) != string(resultRawJSON) {
		t.Errorf("RawJSON not equal expected result")
	}
}

func TestPostHandler(t *testing.T) {
	triggerCount = 0
	expectedRawJSON := json.RawMessage(`{id:"id_test", check: 100}`)

	eventData := &Event{
		Field: "manyfunctions",
	}

	eventManager := NewEventManager()
	eventManager.UsePostHandler(
		func(ctx context.Context, event *Event, result interface{}, err error) {
			triggerCount++
		},
		func(ctx context.Context, event *Event, result interface{}, err error) {
			triggerCount++
		},
		func(ctx context.Context, event *Event, result interface{}, err error) {
			triggerCount++
		},
	)
	eventManager.RegisterField("manyfunctions", testMainHandlers.handler, testMainHandlers.preHandlers, testMainHandlers.postHandlers)

	result, err := eventManager.Run(context.Background(), eventData)
	if err != nil {
		t.Errorf("Next return err %s", err.Error())
	}

	if triggerCount != 5 {
		t.Errorf("Handlers not trigger")
	}

	resultRawJSON := result.Data.(json.RawMessage)
	if string(expectedRawJSON) != string(resultRawJSON) {
		t.Errorf("RawJSON not equal expected result")
	}
}
