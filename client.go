package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tickoalcantara12/stream-app/transport"
)

// Client is the realtime SDK client.
type Client struct {
	cfg Config
	t   transport.Transport

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	handlers handlerStore

	// protect internal flags
	mu        sync.RWMutex
	connected bool
}

// NewClient creates a client with default websocket transport.
func NewClient(cfg Config) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	t := transport.NewWebSocketTransport(cfg.Endpoint, cfg.APIKey)
	return &Client{
		cfg:    cfg,
		t:      t,
		ctx:    ctx,
		cancel: cancel,
	}
}

// NewClientWithTransport creates a client with injected transport (useful for tests).
func NewClientWithTransport(cfg Config, t transport.Transport) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		cfg:    cfg,
		t:      t,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Connect opens the transport and starts background loops.
func (c *Client) Connect() error {
	c.mu.Lock()
	if c.connected {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	if c.t == nil {
		return errors.New("transport is nil")
	}
	if err := c.t.Connect(); err != nil {
		return err
	}

	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	// start receive loop
	c.wg.Add(1)
	go c.receiveLoop()

	// heartbeat loop
	if c.cfg.HeartbeatInterval > 0 {
		c.wg.Add(1)
		go c.heartbeatLoop()
	}

	return nil
}

// receiveLoop reads messages and dispatches events to handlers.
func (c *Client) receiveLoop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		b, err := c.t.Recv()
		if err != nil {
			// transport error -- stop receive loop. reconnection can be added later.
			return
		}

		var evt Event
		if err := json.Unmarshal(b, &evt); err != nil {
			// skip malformed message
			continue
		}

		// dispatch to handlers non-blocking
		for _, h := range c.handlers.list() {
			h := h
			go safeCall(h, evt)
		}
	}
}

// heartbeatLoop sends periodic heartbeat events (optional).
func (c *Client) heartbeatLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.cfg.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			hb := Event{
				ID:       fmt.Sprintf("hb-%d", time.Now().UnixNano()),
				Type:     EventCustom,
				StreamID: "",
				Data:     map[string]string{"type": "heartbeat"},
				At:       time.Now(),
			}
			_ = c.t.Send(marshalOrNil(hb))
		case <-c.ctx.Done():
			return
		}
	}
}

// OnEvent registers an event handler callback.
func (c *Client) OnEvent(fn EventHandler) {
	c.handlers.add(fn)
}

// StartStream signals the backend that this client started a stream and returns the stream id.
func (c *Client) StartStream(meta Metadata) (string, error) {
	// require connection
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()
	if !connected {
		return "", errors.New("not connected")
	}

	id := uuid.NewString()
	evt := Event{
		ID:       uuid.NewString(),
		Type:     EventStreamStart,
		StreamID: id,
		Data:     meta,
		At:       time.Now(),
	}
	b, err := marshalOrErr(evt)
	if err != nil {
		return "", err
	}
	if err := c.t.Send(b); err != nil {
		return "", err
	}
	return id, nil
}

// EndStream signals the backend the stream has ended.
func (c *Client) EndStream(streamID string) error {
	if streamID == "" {
		return errors.New("stream id empty")
	}
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()
	if !connected {
		return errors.New("not connected")
	}
	evt := Event{
		ID:       uuid.NewString(),
		Type:     EventStreamEnd,
		StreamID: streamID,
		Data:     nil,
		At:       time.Now(),
	}
	b, err := marshalOrErr(evt)
	if err != nil {
		return err
	}
	return c.t.Send(b)
}

// SendVideoEvent lets client send a custom stream-related event (viewer join/leave, metadata update).
func (c *Client) SendVideoEvent(streamID string, data interface{}) error {
	if streamID == "" {
		return errors.New("stream id empty")
	}
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()
	if !connected {
		return errors.New("not connected")
	}
	evt := Event{
		ID:       uuid.NewString(),
		Type:     EventCustom,
		StreamID: streamID,
		Data:     data,
		At:       time.Now(),
	}
	b, err := marshalOrErr(evt)
	if err != nil {
		return err
	}
	return c.t.Send(b)
}

// Close stops background loops and closes transport.
func (c *Client) Close() error {
	// cancel context, wait groups, close transport
	c.cancel()
	c.wg.Wait()
	if c.t != nil {
		return c.t.Close()
	}
	return nil
}

// helper to marshal event to JSON and return bytes. return nil if marshal fails.
func marshalOrNil(e Event) []byte {
	b, _ := json.Marshal(e)
	return b
}

// helper to marshal event to JSON and return error if any.
func marshalOrErr(e Event) ([]byte, error) {
	return json.Marshal(e)
}

// safeCall prevents a panic inside an event handler from crashing the recv loop.
func safeCall(fn EventHandler, evt Event) {
	defer func() {
		if r := recover(); r != nil {
			// optional: log or ignore silently
			// fmt.Println("handler panic recovered:", r)
		}
	}()
	fn(evt)
}
