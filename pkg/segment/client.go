package segment

import (
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/analytics-go/v3"
)

var (
	client    analytics.Client
	userId    string
	scope     string
	sessionId = uuid.NewString()
)

const (
	START_STATUS             = "start"
	FAILURE_STATUS           = "failure"
	SUCCESS_STATUS           = "success"
	SESSION_ID_PROPERTY_NAME = "sessionId"
	ERROR_PROPERTY_NAME      = "error"
	STATUS_PROPERTY_NAME     = "status"
	SCOPE_PROPERTY_NAME      = "scope"
	ORG_TARIAT_NAME          = "orgName"
	PROD_WRITE_KEY           = ""
	DEV_WRITE_KEY            = "FPPzr8mdiYq9Ry2YOEVFN751DvSdwwUZ"
)

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

func (event *EventHandler) Failure(err error) error {
	event.Set(ERROR_PROPERTY_NAME, err)
	return event.enqueueWithStatus(FAILURE_STATUS)
}

func (event *EventHandler) Success() error {
	return event.enqueueWithStatus(SUCCESS_STATUS)
}

func (event *EventHandler) enqueueWithStatus(status string) error {
	event.Properties.Set(STATUS_PROPERTY_NAME, status)
	return client.Enqueue(event.Track)
}

func Init(appName, environment, release string) error {
	var err error

	config := analytics.Config{
		BatchSize: 1,
		Interval:  5 * time.Second,

		DefaultContext: &analytics.Context{
			App: analytics.AppInfo{
				Name:    appName,
				Version: release,
			},
		},
	}

	writeKey := PROD_WRITE_KEY

	if environment == "dev" {
		writeKey = DEV_WRITE_KEY
	}

	if client, err = analytics.NewWithConfig(writeKey, config); err != nil {
		return err
	}

	return nil
}

func Close() error {
	return client.Close()
}

func NewUser(email string, org string) error {
	var err error

	user := analytics.Identify{
		UserId: email,
		Traits: analytics.NewTraits().SetEmail(email).Set(ORG_TARIAT_NAME, org),
	}

	orgGroup := analytics.Group{
		GroupId: org,
		UserId:  user.UserId,
		Traits:  analytics.NewTraits().SetName(org),
	}

	if err = client.Enqueue(user); err != nil {
		return err
	}

	if err = client.Enqueue(orgGroup); err != nil {
		return err
	}

	SetUser(email)
	return nil
}

func SetUser(email string) {
	userId = email
}

func SetScope(name string) {
	scope = name
}

func GetScope() string {
	return scope
}

func SetSessionId(id string) {
	sessionId = id
}
