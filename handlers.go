package stream

import "sync"

type EventHandler func(Event)

type handlerStore struct {
	mu       sync.RWMutex
	handlers []EventHandler
}

func (h *handlerStore) add(fn EventHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers = append(h.handlers, fn)
}

func (h *handlerStore) list() []EventHandler {
	h.mu.RLock()
	defer h.mu.RUnlock()

	cp := make([]EventHandler, len(h.handlers))
	copy(cp, h.handlers)
	return cp
}
