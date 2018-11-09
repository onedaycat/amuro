package appsync

import (
	"context"
	"testing"
)

func TestRunMiddlewares(t *testing.T) {
	eventManager := NewEventManager()

	input := &Event{
		Field:    "test_middle",
		Args:     nil,
		Identity: nil,
	}

	runTime := 0

	eventManager.RegisterMiddleware(func(ctx context.Context, event *Event) (interface{}, error) {
		runTime++
		return nil, nil
	})

	eventManager.RegisterMiddleware(func(ctx context.Context, event *Event) (interface{}, error) {
		runTime++
		return nil, nil
	})

	eventManager.RegisterField("test_middle", func(ctx context.Context, event *Event) (interface{}, error) {
		runTime++
		return "i'm here", nil
	})

	result, err := eventManager.DefaultHandler(context.Background(), input)
	if err != nil {
		t.Error(err)
	}

	r := result.(*Result)
	if r.Error != nil {
		t.Error(r.Error)
	}
	if r.Data.(string) != "i'm here" {
		t.Error("result not equal `i'm here`")
	}

	if runTime != 3 {
		t.Error("middlewares or hanlder not trigger")
	}
}
