package transport

import (
	"errors"
	"time"
)

const (
	mockClosed = "mock closed"
)

type MockTransport struct {
	SendCh chan []byte
	RecvCh chan []byte
	Closed bool
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		SendCh: make(chan []byte, 10),
		RecvCh: make(chan []byte, 10),
	}
}

func (m *MockTransport) Connect() error {
	if m.Closed {
		return errors.New(mockClosed)
	}
	return nil
}

func (m *MockTransport) Send(b []byte) error {
	if m.Closed {
		return errors.New(mockClosed)
	}
	m.SendCh <- b
	return nil
}

func (m *MockTransport) Recv() ([]byte, error) {
	if m.Closed {
		return nil, errors.New("mock closed")
	}

	select {
	case msg := <-m.RecvCh:
		return msg, nil
	case <-time.After(50 * time.Millisecond):
		// simulate timeout → receiveLoop will exit safely
		return nil, errors.New("recv timeout")
	}
}

func (m *MockTransport) Close() error {
	m.Closed = true
	return nil
}
