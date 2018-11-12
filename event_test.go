package appsync

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/onedaycat/errors"
)

func TestRegisterHandlerAndRun(t *testing.T) {
	triggerCount := 0
	rawJSON := json.RawMessage(`{id:"id_test", check: 100}`)
	handlers := []EventHandler{
		func(ctx EventContext, event *Event) (interface{}, error) {
			triggerCount++
			return nil, nil
		},
		func(ctx EventContext, event *Event) (interface{}, error) {
			triggerCount++
			return event.Args, nil
		},
		func(ctx EventContext, event *Event) (interface{}, error) {
			triggerCount++
			return nil, nil
		},
	}

	eventData := &Event{
		Field: "manyfunctions",
		Args:  rawJSON,
	}

	eventManager := NewEventManager()
	eventManager.RegisterField("manyfunctions", handlers)

	result, err := eventManager.Run(context.Background(), eventData)
	if err != nil {
		t.Errorf("Next return err %s", err.Error())
	}

	if triggerCount != len(handlers) {
		t.Errorf("Function in handlers not trigger")
	}

	resultRawJSON := result.Data.(json.RawMessage)
	if string(rawJSON) != string(resultRawJSON) {
		t.Errorf("RawJSON not equal expected result")
	}
}

func TestRegisterEmptyHandlers(t *testing.T) {
	emptyHandlers := []EventHandler{}
	expectedError := errors.InternalErrorf("UNABLE_REGISTER_FIELD", "Handlers cant be empty")

	eventManager := NewEventManager()

	err := eventManager.RegisterField("emptyfunctions", emptyHandlers)
	if err.Error() != expectedError.Error() {
		t.Errorf("Error not equal exepected")
	}
}
