package custom_sentry

import (
	"fmt"
	"strconv"

	"github.com/getsentry/sentry-go"
	"groundcover.com/pkg/auth"
)

const (
	LOGIN_EVENT_MESSAGE      = "login"
	DEPLOYMENT_EVENT_MESSAGE = "deployment"
	ORGANIZATION_TAG         = "organization"
	VERSION_TAG              = "version"
	UPGRADE_TAG              = "upgrade"
	NUMBER_OF_NODES_TAG      = "number_of_nodes"
)

func CaptureLoginEvent(customClaims *auth.CustomClaims) {
	event := sentry.NewEvent()

	event.Message = LOGIN_EVENT_MESSAGE
	event.Tags = map[string]string{
		ORGANIZATION_TAG: customClaims.Org,
	}
	event.User = sentry.User{
		Email:    customClaims.Email,
		Username: customClaims.Email,
	}

	sentry.CaptureEvent(event)
}

func CaptureDeploymentEvent(customClaims *auth.CustomClaims, isUpgrade bool, version string, numberOfNodes int) {
	event := sentry.NewEvent()

	event.Message = DEPLOYMENT_EVENT_MESSAGE
	event.Tags = map[string]string{
		ORGANIZATION_TAG:    customClaims.Org,
		VERSION_TAG:         version,
		UPGRADE_TAG:         strconv.FormatBool(isUpgrade),
		NUMBER_OF_NODES_TAG: fmt.Sprintf("%d", numberOfNodes),
	}
	event.User = sentry.User{
		Email:    customClaims.Email,
		Username: customClaims.Email,
	}

	sentry.CaptureEvent(event)
}
