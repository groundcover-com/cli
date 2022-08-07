package sentry_test

import (
	"sync"
	"time"

	. "github.com/getsentry/sentry-go"
)

type TransportMock struct {
	mu        sync.Mutex
	events    []*Event
	lastEvent *Event
}

func (t *TransportMock) Configure(options ClientOptions) {}

func (t *TransportMock) SendEvent(event *Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
	t.lastEvent = event
}

func (t *TransportMock) Flush(timeout time.Duration) bool {
	return true
}

func (t *TransportMock) Events() []*Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.events
}
