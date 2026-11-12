package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSEMessage represents an event payload sent over Server-Sent Events.
type SSEMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// EventBroker manages active SSE client subscriptions keyed by batch ID.
type EventBroker struct {
	mu      sync.RWMutex
	clients map[string]map[chan SSEMessage]struct{}
}

// NewEventBroker initializes the SSE broker.
func NewEventBroker() *EventBroker {
	return &EventBroker{
		clients: make(map[string]map[chan SSEMessage]struct{}),
	}
}

// Subscribe registers a new SSE channel for a batch ID.
func (b *EventBroker) Subscribe(batchID string) chan SSEMessage {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan SSEMessage, 32)
	if _, ok := b.clients[batchID]; !ok {
		b.clients[batchID] = make(map[chan SSEMessage]struct{})
	}
	b.clients[batchID][ch] = struct{}{}
	return ch
}

// Unsubscribe removes an SSE channel.
func (b *EventBroker) Unsubscribe(batchID string, ch chan SSEMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.clients[batchID]; ok {
		delete(subs, ch)
		close(ch)
		if len(subs) == 0 {
			delete(b.clients, batchID)
		}
	}
}

// Broadcast sends an event to all subscribers listening to a batch ID.
func (b *EventBroker) Broadcast(batchID, event string, data interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	subs, ok := b.clients[batchID]
	if !ok {
		return
	}

	msg := SSEMessage{Event: event, Data: data}
	for ch := range subs {
		select {
		case ch <- msg:
		default:
		}
	}
}

// ServeSSE handles the HTTP connection for Server-Sent Events.
func (b *EventBroker) ServeSSE(w http.ResponseWriter, r *http.Request, batchID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := b.Subscribe(batchID)
	defer b.Unsubscribe(batchID, ch)

	// Send initial ping
	fmt.Fprintf(w, "event: ping\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			dataBytes, err := json.Marshal(msg.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", msg.Event, string(dataBytes))
			flusher.Flush()
		}
	}
}
