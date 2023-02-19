package sentry_test

import (
	"fmt"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	sentry_utils "groundcover.com/pkg/sentry"
)

type SentryClientTestSuite struct {
	suite.Suite
	Transport *TransportMock
}

func (suite *SentryClientTestSuite) SetupSuite() {
	suite.Transport = &TransportMock{}

	clientOptions := sentry.ClientOptions{
		Dsn:          "http://whatever@really.com/1337",
		Transport:    suite.Transport,
		Integrations: func(i []sentry.Integration) []sentry.Integration { return []sentry.Integration{} },
	}

	client, _ := sentry.NewClient(clientOptions)
	sentry.CurrentHub().BindClient(client)
}

func (suite *SentryClientTestSuite) TearDownSuite() {}

func TestSentryClientSuite(t *testing.T) {
	suite.Run(t, &SentryClientTestSuite{})
}

func (suite *SentryClientTestSuite) TestGetSentryClientOptionsProd() {
	//prepare
	appName := "cli"
	version := "1.0.0"
	environment := "prod"

	//act
	clientOptions := sentry_utils.GetSentryClientOptions(appName, environment, version)

	// assert
	expect := sentry.ClientOptions{
		MaxBreadcrumbs: 10,
		Release:        fmt.Sprintf("%s@%s", appName, version),
		Environment:    environment,
		Dsn:            sentry_utils.PROD_DSN,
	}

	suite.Equal(expect, clientOptions)
}

func (suite *SentryClientTestSuite) TestGetSentryClientOptionsDev() {
	//prepare
	appName := "cli"
	version := "1.0.0"
	environment := "dev"

	//act
	clientOptions := sentry_utils.GetSentryClientOptions(appName, environment, version)

	// assert
	expect := sentry.ClientOptions{
		MaxBreadcrumbs: 10,
		Release:        fmt.Sprintf("%s@%s", appName, version),
		Environment:    environment,
		Dsn:            sentry_utils.DEV_DSN,
	}

	suite.Equal(expect, clientOptions)
}

func (suite *SentryClientTestSuite) TestSetOnCurrentScopeSuccess() {
	//prepare
	level := sentry.LevelWarning
	tagName := uuid.New().String()
	tagValue := uuid.New().String()
	transaction := uuid.New().String()

	user := sentry.User{
		Email:    uuid.New().String(),
		Username: uuid.New().String(),
	}

	//act
	sentry_utils.SetUserOnCurrentScope(user)
	sentry_utils.SetLevelOnCurrentScope(level)
	sentry_utils.SetTagOnCurrentScope(tagName, tagValue)
	sentry_utils.SetTransactionOnCurrentScope(transaction)
	sentry.CaptureMessage("set on scope")

	// assert
	event := suite.Transport.lastEvent

	suite.Equal(user, event.User)
	suite.Equal(transaction, event.Transaction)
	suite.Equal(map[string]string{tagName: tagValue}, event.Tags)
}
