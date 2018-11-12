package appsync

import (
	"testing"
)

func TestEventContextNext(t *testing.T) {
	triggerCount := 0
	eventData := &Event{}
	handlers := []EventHandler{
		func(ctx EventContext, event *Event) (interface{}, error) {
			triggerCount++
			return nil, nil
		},
		func(ctx EventContext, event *Event) (interface{}, error) {
			triggerCount++
			return nil, nil
		},
	}

	eventContext := NewEventContext(handlers, eventData)
	_, err := eventContext.Next()

	if err != nil {
		t.Errorf("Next return err %s", err.Error())
	}

	if triggerCount != len(handlers) {
		t.Errorf("Function in handlers not trigger")
	}
}
