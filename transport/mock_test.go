package transport

import (
	"testing"
	"time"
)

func TestMockTransport_SendRecv(t *testing.T) {
	m := NewMockTransport()

	if err := m.Connect(); err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	payload := []byte("hello")
	if err := m.Send(payload); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	select {
	case got := <-m.SendCh:
		if string(got) != "hello" {
			t.Fatalf("unexpected send payload: %s", string(got))
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting send")
	}

	// simulate server message
	m.RecvCh <- []byte("server-msg")
	b, err := m.Recv()
	if err != nil {
		t.Fatalf("recv failed: %v", err)
	}
	if string(b) != "server-msg" {
		t.Fatalf("unexpected recv payload: %s", string(b))
	}

	if err := m.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
}
