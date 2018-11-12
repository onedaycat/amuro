package appsync

import (
	"context"
	"encoding/json"
	"testing"
)

var triggerCount = 0
var testPostHandlers = []EventHandler{
	func(ctx context.Context, event *Event) (interface{}, error) {
		triggerCount++
		return nil, nil
	},
	func(ctx context.Context, event *Event) (interface{}, error) {
		triggerCount++
		return nil, nil
	},
}
var testPreHandlers = []EventHandler{
	func(ctx context.Context, event *Event) (interface{}, error) {
		triggerCount++
		return nil, nil
	},
}
var testHandlers = []EventHandler{
	func(ctx context.Context, event *Event) (interface{}, error) {
		triggerCount++
		return nil, nil
	},
	func(ctx context.Context, event *Event) (interface{}, error) {
		triggerCount++
		return event.Args, nil
	},
	func(ctx context.Context, event *Event) (interface{}, error) {
		triggerCount++
		return nil, nil
	},
}

func TestRegisterHandlerAndRun(t *testing.T) {
	triggerCount = 0
	rawJSON := json.RawMessage(`{id:"id_test", check: 100}`)

	eventData := &Event{
		Field: "manyfunctions",
		Args:  rawJSON,
	}

	eventManager := NewEventManager()
	eventManager.RegisterFields("manyfunctions", testHandlers)

	result, err := eventManager.Run(context.Background(), eventData)
	if err != nil {
		t.Errorf("Next return err %s", err.Error())
	}

	if triggerCount != len(testHandlers) {
		t.Errorf("Function in handlers not trigger")
	}

	resultRawJSON := result.Data.(json.RawMessage)
	if string(rawJSON) != string(resultRawJSON) {
		t.Errorf("RawJSON not equal expected result")
	}
}

func TestRegisterPreHandlerAndRun(t *testing.T) {
	triggerCount = 0
	rawJSON := json.RawMessage(`{id:"id_test", check: 100}`)

	eventData := &Event{
		Field: "manyfunctions",
		Args:  rawJSON,
	}

	eventManager := NewEventManager()
	eventManager.RegisterPreFunction(testPreHandlers)
	eventManager.RegisterFields("manyfunctions", testHandlers)

	result, err := eventManager.Run(context.Background(), eventData)
	if err != nil {
		t.Errorf("Next return err %s", err.Error())
	}

	if triggerCount != 4 {
		t.Errorf("Function in handlers not trigger")
	}

	resultRawJSON := result.Data.(json.RawMessage)
	if string(rawJSON) != string(resultRawJSON) {
		t.Errorf("RawJSON not equal expected result")
	}
}

func TestRegisterPostHandlerAndRun(t *testing.T) {
	triggerCount = 0
	rawJSON := json.RawMessage(`{id:"id_test", check: 100}`)

	eventData := &Event{
		Field: "manyfunctions",
		Args:  rawJSON,
	}

	eventManager := NewEventManager()
	eventManager.RegisterPostFunction(testPostHandlers)
	eventManager.RegisterFields("manyfunctions", testHandlers)

	result, err := eventManager.Run(context.Background(), eventData)
	if err != nil {
		t.Errorf("Next return err %s", err.Error())
	}

	if triggerCount != 5 {
		t.Errorf("Function in handlers not trigger")
	}

	resultRawJSON := result.Data.(json.RawMessage)
	if string(rawJSON) != string(resultRawJSON) {
		t.Errorf("RawJSON not equal expected result")
	}
}

func TestRegisterEmptyHandlers(t *testing.T) {
	emptyHandlers := []EventHandler{}
	eventData := &Event{
		Field: "emptyHandler",
	}

	eventManager := NewEventManager()
	eventManager.RegisterFields("emptyHandler", emptyHandlers)
	result, err := eventManager.Run(context.Background(), eventData)

	if err.Error() != "FIELD_NOT_FOUND: Not found handler on field emptyHandler" {
		t.Errorf("Field not found should be error")
	}

	if result != nil {
		t.Errorf("Result should be empty")
	}

}
