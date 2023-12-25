package multiplex

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

func TestMessageMux_serve(t *testing.T) {
	// arrange
	type redisMessage struct {
		ctx     context.Context
		payload []byte
		channel string
	}
	recorder := &bytes.Buffer{}

	mux := NewMessageMux(func(message *redisMessage) (string, error) {
		return message.channel, nil
	})

	mux.
		RegisterFunc("hello", func(dto *redisMessage) error {
			fmt.Fprintf(recorder, "topic=%v, payload=%v", dto.channel, string(dto.payload))
			return nil
		}).
		RegisterFunc("foo", func(dto *redisMessage) error {
			fmt.Fprintf(recorder, "topic=%v, payload=%v", dto.channel, string(dto.payload))
			return nil
		})

	// expect
	expected := `topic=hello, payload={"data":"world"}`

	// action
	dto := &redisMessage{
		ctx:     context.Background(),
		payload: []byte(`{"data":"world"}`),
		channel: "hello",
	}
	err := mux.serve(dto)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// assert
	if expected != recorder.String() {
		t.Errorf("serve(): %v, but want: %v", recorder.String(), expected)
	}
}
