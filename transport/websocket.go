package transport

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type WebSocketTransport struct {
	url    string
	apiKey string
	conn   *websocket.Conn
}

func NewWebSocketTransport(url, key string) *WebSocketTransport {
	return &WebSocketTransport{url: url, apiKey: key}
}

func (t *WebSocketTransport) Connect() error {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+t.apiKey)

	c, _, err := websocket.DefaultDialer.Dial(t.url, h)
	if err != nil {
		return err
	}
	t.conn = c
	return nil
}

func (t *WebSocketTransport) Send(b []byte) error {
	return t.conn.WriteMessage(websocket.TextMessage, b)
}

func (t *WebSocketTransport) Recv() ([]byte, error) {
	_, msg, err := t.conn.ReadMessage()
	return msg, err
}

func (t *WebSocketTransport) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}
