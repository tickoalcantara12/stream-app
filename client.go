package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tickoalcantara12/stream-app/transport"
)

type Client struct {
	cfg Config
	t   transport.Transport

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	handlers handlerStore
}

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

func (c *Client) Connect() error {
	if err := c.t.Connect(); err != nil {
		return err
	}

	// Receive loop
	c.wg.Add(1)
	go c.receiveLoop()

	// Heartbeat
	if c.cfg.HeartbeatInterval > 0 {
		c.wg.Add(1)
		go c.heartbeatLoop()
	}

	return nil
}

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
			return
		}

		var evt Event
		if err := json.Unmarshal(b, &evt); err != nil {
			continue
		}

		for _, h := range c.handlers.list() {
			go h(evt)
		}
	}
}

func (c *Client) heartbeatLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.cfg.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hb := Event{
				ID:   "hb",
				Type: "heartbeat",
				At:   time.Now(),
			}
			b, _ := json.Marshal(hb)
			c.t.Send(b)
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Client) OnEvent(fn EventHandler) {
	c.handlers.add(fn)
}

func (c *Client) StartStream(name string, meta Metadata) (string, error) {
	id := fmt.Sprintf("stream-%d", time.Now().UnixNano())

	evt := Event{
		ID:       id,
		Type:     EventStreamStart,
		StreamID: id,
		Data:     meta,
		At:       time.Now(),
	}

	b, _ := json.Marshal(evt)
	if err := c.t.Send(b); err != nil {
		return "", err
	}

	return id, nil
}

func (c *Client) EndStream(id string) error {
	if id == "" {
		return errors.New("stream id is empty")
	}

	evt := Event{
		ID:       id + "-end",
		Type:     EventStreamEnd,
		StreamID: id,
		At:       time.Now(),
	}

	b, _ := json.Marshal(evt)
	return c.t.Send(b)
}

func (c *Client) SendEvent(streamID string, data interface{}) error {
	evt := Event{
		ID:       fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Type:     EventCustom,
		StreamID: streamID,
		Data:     data,
		At:       time.Now(),
	}

	b, _ := json.Marshal(evt)
	return c.t.Send(b)
}

func (c *Client) Close() error {
	c.cancel()
	c.wg.Wait()
	return c.t.Close()
}
