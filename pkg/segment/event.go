package segment

import (
	"github.com/google/uuid"
	"github.com/segmentio/analytics-go/v3"
)

const (
	ABORT_STATUS             = "abort"
	START_STATUS             = "start"
	FAILURE_STATUS           = "failure"
	SUCCESS_STATUS           = "success"
	SCOPE_PROPERTY_NAME      = "scope"
	ERROR_PROPERTY_NAME      = "error"
	STATUS_PROPERTY_NAME     = "status"
	SESSION_ID_PROPERTY_NAME = "sessionId"
)

var (
	scope     string
	sessionId = uuid.NewString()
)

func SetScope(name string) {
	scope = name
}

func GetScope() string {
	return scope
}

func SetSessionId(id string) {
	sessionId = id
}

type EventHandler struct {
	analytics.Track
}

func NewEvent(name string) *EventHandler {
	event := &EventHandler{}
	event.Event = name
	event.UserId = userId
	if userId == "" {
		event.AnonymousId = sessionId
	}
	event.Properties = analytics.NewProperties()
	event.Properties.Set(SCOPE_PROPERTY_NAME, scope)
	event.Properties.Set(SESSION_ID_PROPERTY_NAME, sessionId)

	return event
}

func (event *EventHandler) Set(name string, value interface{}) analytics.Properties {
	event.Properties.Set(name, value)
	return event.Properties
}

func (event *EventHandler) Start() error {
	return event.enqueueWithStatus(START_STATUS)
}

func (event *EventHandler) Abort() error {
	return event.enqueueWithStatus(ABORT_STATUS)
}

func (event *EventHandler) Failure(err error) error {
	event.Set(ERROR_PROPERTY_NAME, err.Error())
	return event.enqueueWithStatus(FAILURE_STATUS)
}

func (event *EventHandler) Success() error {
	return event.enqueueWithStatus(SUCCESS_STATUS)
}

func (event *EventHandler) enqueueWithStatus(status string) error {
	event.Properties.Set(STATUS_PROPERTY_NAME, status)
	return client.Enqueue(event.Track)
}
