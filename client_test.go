package stream

import (
	"encoding/json"
	"github.com/tickoalcantara12/stream-app/transport"
	"testing"
	"time"
)

func TestClientWithMockTransport_SendAndReceive(t *testing.T) {
	cfg := DefaultConfig()
	mt := transport.NewMockTransport()
	c := NewClientWithTransport(cfg, mt)

	// connect (mock does nothing)
	if err := c.Connect(); err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	received := make(chan Event, 1)
	c.OnEvent(func(e Event) {
		received <- e
	})

	// simulate server sending an event
	go func() {
		msg := Event{
			ID:       "srv-1",
			Type:     EventCustom,
			StreamID: "s1",
			Data:     map[string]string{"msg": "hi"},
			At:       time.Now(),
		}
		b, _ := json.Marshal(msg)
		mt.RecvCh <- b
	}()

	select {
	case ev := <-received:
		if ev.ID != "srv-1" {
			t.Fatalf("unexpected event id: %s", ev.ID)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting event")
	}

	// test send
	if err := c.SendVideoEvent("s1", map[string]string{"k": "v"}); err != nil {
		t.Fatalf("send error: %v", err)
	}

	select {
	case sb := <-mt.SendCh:
		var evt Event
		if err := json.Unmarshal(sb, &evt); err != nil {
			t.Fatalf("unmarshal payload error: %v", err)
		}
		if evt.StreamID != "s1" {
			t.Fatalf("unexpected stream id in send: %s", evt.StreamID)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting sendCh")
	}

	_ = c.Close()
}
